package main

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"io/fs"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

//go:embed public/*
var public embed.FS

var l = slog.Default()

func main() {
	godotenv.Load()
	gin.SetMode(gin.ReleaseMode)
	
	db := getDB()
	if db == nil {
		l.Error("Failed to get DB")
		os.Exit(1)
	}

	defer db.Close()
	
	r := gin.Default()

	ao := os.Getenv("ALLOWED_ORIGINS")
	var allowOrigins []string
	aoTokens := strings.Split(ao, ",")
	for _, token := range aoTokens {
		allowOrigins = append(allowOrigins, strings.TrimSpace(token))
	}
	
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	r.GET("/_", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/_/:slug", func(c *gin.Context) {
		db := c.MustGet("db").(*sql.DB)
		slug := c.Param("slug")
		var target string

		if err := db.QueryRow(`
			SELECT target FROM urls_v1 WHERE slug = ?
		`, slug).Scan(&target); err != nil {
			switch err {
				case sql.ErrNoRows:
					data, err := public.ReadFile("public/404-slug.html")
					if err != nil {
						c.String(http.StatusNotFound, "Invalid slug!")
						return
					}
					c.Data(http.StatusNotFound, "text/html; charset=utf-8", data)
					return
				default:
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "database failure",
					})
					l.Error(err.Error())
					return	
			}
		}

		c.Redirect(http.StatusFound, target)
	})

	r.POST("/generate", func(c *gin.Context) {
		db := c.MustGet("db").(*sql.DB)

		var payload struct {
			URL string `json:"url"`
		}

		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid payload",
			})
			l.Error(err.Error())
			return
		}

		if len(payload.URL) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "url is empty",
			})
			l.Warn("payload url is empty")
			return
		}
		
		if !validateURL(payload.URL) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid url",
			})
			l.Warn("payload url is invalid")
			return
		}

		slug := GenerateRandomChars(10)
		url := ""

		if len(slug) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to shorten link",
			})
			l.Warn("slug is empty")
			return
		}
		
		scheme := "https"
		if c.Request.TLS == nil {
		    scheme = "http"
		}
		
		host := scheme + "://" + c.Request.Host

		_, err := db.Exec(`
			INSERT INTO URLs_v1 
			(url_id, slug, target) 
			VALUES (?, ?, ?)
		`, time.Now().Unix(), slug, payload.URL);

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database failure",
			})
			l.Error(err.Error())
			return
		}

		url = host + "/_/" + slug

		c.JSON(http.StatusOK, gin.H{
			"data": url,
		})
	})

	{
		// publicFs, _ := fs.Sub(public, "public")

		// Serve index.html at root
		r.GET("/", func(c *gin.Context) {
			data, err := public.ReadFile("public/index.html")
			if err != nil {
				c.String(http.StatusNotFound, "Not found")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})

		assetsFs, _ := fs.Sub(public, "public/assets")
		r.StaticFS("/assets", http.FS(assetsFs))
		
		r.NoRoute(func(c *gin.Context) {
			data, err := public.ReadFile("public/404.html")
			if err != nil {
				c.String(http.StatusNotFound, "You might be lost")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	r.Run(":" + port)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomChars(n int) string {
	result := make([]byte, n)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[index.Int64()]
	}
	return string(result)
}

func validateURL(raw string) bool {
    u, err := url.ParseRequestURI(raw)
    return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func getDB() *sql.DB {
	var db *sql.DB
	url := os.Getenv("TDB_URL") + "?authToken=" + os.Getenv("TDB_TOKEN")

	db, err := sql.Open("libsql", url)
	if err != nil {
		l.Error(err.Error())
		return nil
	}

	if err := db.Ping(); err != nil {
		l.Error(err.Error())
		return nil
	}
	return db
}