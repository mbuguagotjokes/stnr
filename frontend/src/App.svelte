<script lang="ts">
  import { onMount } from "svelte";
  import { Spring } from 'svelte/motion';
  import favicon from "./assets/favicon.svg"
  import ogImage from "./assets/og-image.png"
  import { fade, slide } from "svelte/transition";
  
  let coords = new Spring({ x: 50, y: 50 }, {
  	stiffness: 0.01,
  	damping: 0.25
  });
  
  let size = new Spring(10);

  let from = $state("")
  let to = $state("")
  let error = $state("")
  let msg = $state("")
  let submiting = $state(false)

  let loading = $state(true)
  onMount(async() => {
    setTimeout(() => {
      loading = false
    }, 3000)
  })
</script>

<svelte:head>
    <link rel="icon" type="image/svg+xml" href={favicon}>
    <meta property="og:type"        content="website" />
    <meta property="og:site_name"   content="STNR - Simple link shortener" />
    <meta property="og:title"       content="STNR" />
    <meta property="og:description" content="A very very simple link shortener." />
    <meta property="og:url"         content="https://stnr.mbuguaaaaaa.xyz" />
    <meta property="og:image"       content={ogImage} />
    <meta property="og:image:width"  content="1200" />
    <meta property="og:image:height" content="630" />
    <meta property="og:image:alt"   content="STNR — A very very simple link shortener" />
    <meta property="og:locale"      content="en_US" />
    <meta name="twitter:card"        content="summary_large_image" />
    <meta name="twitter:title"       content="STNR" />
    <meta name="twitter:description" content="A very very simple link shortener." />
    <meta name="twitter:image"       content={ogImage} />
    <meta name="twitter:image:alt"   content="STNR — A very very simple link shortener" />
    <meta name="theme-color" content="#F4442E" />
    <link rel="canonical" href="https://stnr.mbuguaaaaaa.xyz" />
    <meta name="description" content="A very very simple link shortener." />
</svelte:head>

<!-- <svg
	onmousemove={(e) => {
		coords.from = { x: e.clientX, y: e.clientY };
	}}
	onmousedown={() => (size.from = 30)}
	onmouseup={() => (size.from = 10)}
	role="presentation"
>
	<circle
		cx={coords.current.x}
		cy={coords.current.y}
		r={size.current}
	/>
</svg> -->

{#if loading}
    <div 
        class="loading"
        transition:slide
    >
        <h1 class="cprimary heartbeat">STNR</h1>
        <p class="csecondary">Shortener, simple link shortener</p>
    </div>
{:else}
    <div class="shorten">
        <section class="head">
            <h1 class="cprimary">STNR</h1>
            <p class="csecondary">Shortener, simple link shortener</p>
        </section>
        <section class="inp">
            <label for="url">Enter the link you want to shorten (must start with either <b>http://</b> or <b>https://)</b></label>
            <input type="text" name="url" id="url" bind:value={from} placeholder="https://mbuguaaaaaa.xyz">
        </section>
        {#if to}
        <section class="to" transition:fade>
            <a href={to} class="to">{to}</a>
        </section>
        {/if}
        {#if error}
        <section class="error" transition:slide>
            <p
                style="text-align: center; color: red;"
            >{error}</p>
        </section>
        {/if}
        {#if msg}
        <section class="msg" transition:slide>
            <p
                style="text-align: center; color: green;"
            >{msg}</p>
        </section>
        {/if}
        <section class="actions">
            {#if to}
                <!-- class="bcsecondary" -->
            <button
                transition:fade
                onclick={async() => {
                  try {
                      await navigator.clipboard.writeText(to)
                      msg = "Copied to clipboard succesfully!"
                      setTimeout(() => {
                        msg = ""
                      }, 3000)
                  } catch (error) {
                      const textarea = document.createElement("textarea");
                      textarea.value = to;
                      
                      textarea.style.position = "fixed";
                      textarea.style.left = "-9999px";
                      
                      document.body.appendChild(textarea);
                      
                      textarea.focus();
                      textarea.select();
                      
                      const successful = document.execCommand("copy");
                      
                      document.body.removeChild(textarea);
                      
                      if (!successful) {
                        error = "Failed to copy to clipboard"
                        setTimeout(() => {
                          error = ""
                        }, 3000)
                        return
                      }
                      msg = "Copied to clipboard succesfully!"
                      setTimeout(() => {
                        msg = ""
                      }, 3000)
                  }
                }}
            >
                Copy Link
            </button>
            {/if}
            <button
                class="bcsecondary"
                disabled={submiting}
                onclick={async() => {
                  submiting = true
                  try {
                    if (from.trim() == "") {
                      error = "Please enter a url in the input field"
                      setTimeout(() => {
                        error = ""
                      }, 3000)
                      return
                    }
                    
                    const sl = await fetch("/generate", {
                      method: "POST",
                      headers: {
                        "Content-Type": "application/json"
                      },
                      body: JSON.stringify({ url: from })
                    })

                    let response = await sl.json()

                    if (sl.ok) {
                      to = response.data
                    } else {
                      error = response.error
                      setTimeout(() => {
                        error = ""
                      }, 3000)
                    }
                  } catch (e) {
                    error = "Failed to shorten link, please try again"
                    setTimeout(() => {
                      error = ""
                    }, 3000)
                  } finally {
                    submiting = false
                  }
                }}
            >
                Shorten link
            </button>
        </section>
    </div>
{/if}

<style>
    /*svg {
		position: absolute;
		width: 100%;
		height: 100%;
		left: 0;
		top: 0;
	}
   
	circle {
		pointer-events: none;
		fill: var(--secondary);
	}*/
	
    .loading {
        width: 100%;
        height: 100vh;
        overflow: hidden;
        display: flex;
        flex-direction: column;
        gap: 8px;
        justify-content: center;
        align-items: center;
        cursor: none;
    }

    .loading > h1 {
        font-size: clamp(8vw, 1em + 10vw, 15vw);
    }

    .shorten {
        width: 100%;
        height: 100vh;
        padding: 16px;
        
        display: flex;
        flex-direction: column;
        gap: 8px;
        justify-content: start;
        align-items: center;
    }

    .head {
        display: flex;
        flex-direction: column;
        gap: 8px;
        justify-content: center;
        align-items: center;
    }

    .head > h1 {
        font-size: clamp(8vw,  10vw, 12vw);
        margin: .2em 0;
    }

    .inp {
        display: flex;
        flex-direction: column;
        gap: 16px;
        margin: 0 16px;
    }

    .actions {
        display: flex;
        gap: 8px;
        justify-content: center;
        align-items: center;
    }
</style>