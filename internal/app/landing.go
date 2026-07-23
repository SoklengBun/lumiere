package app

import (
	"bytes"
	"html/template"
)

var landingPage = template.Must(template.New("landing").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="theme-color" content="#10101b">
  <title>Lumiere — music, beautifully found</title>
  <style>
    :root {
      color-scheme: dark;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #10101b;
      color: #f9f7ff;
    }
    * { box-sizing: border-box; }
    body { margin: 0; min-height: 100vh; overflow: hidden; background: #10101b; }
    .orb { position: fixed; border-radius: 999px; filter: blur(12px); opacity: .55; pointer-events: none; }
    .orb-one { width: 28rem; height: 28rem; top: -12rem; right: -6rem; background: #c56cff; }
    .orb-two { width: 24rem; height: 24rem; bottom: -12rem; left: -7rem; background: #635bff; }
    .shell { position: relative; display: grid; place-items: center; min-height: 100vh; padding: 2rem; }
    .card { width: min(100%, 680px); padding: clamp(2rem, 7vw, 5rem); border: 1px solid rgba(255,255,255,.14); border-radius: 32px; background: rgba(24, 23, 39, .72); box-shadow: 0 30px 90px rgba(0,0,0,.38), inset 0 1px rgba(255,255,255,.08); backdrop-filter: blur(24px); text-align: center; }
    .mark { display: inline-grid; place-items: center; width: 64px; height: 64px; margin-bottom: 1.75rem; border-radius: 20px; background: linear-gradient(135deg, #f0a6ff, #7c72ff); box-shadow: 0 12px 32px rgba(164, 112, 255, .35); }
    .mark svg { width: 32px; height: 32px; fill: #fff; }
    .eyebrow { margin: 0 0 .75rem; color: #c9bfff; font-size: .78rem; font-weight: 700; letter-spacing: .18em; text-transform: uppercase; }
    h1 { margin: 0; font-size: clamp(2.5rem, 8vw, 5.25rem); line-height: .98; letter-spacing: -.065em; }
    .gradient { background: linear-gradient(105deg, #fff 15%, #d4b6ff 75%, #a9a2ff); -webkit-background-clip: text; background-clip: text; color: transparent; }
    .copy { max-width: 30rem; margin: 1.5rem auto 2.25rem; color: #b9b6ca; font-size: 1.08rem; line-height: 1.7; }
    .button { display: inline-flex; align-items: center; gap: .65rem; padding: .95rem 1.35rem; border-radius: 999px; background: #f7f4ff; color: #201b36; font-weight: 800; text-decoration: none; transition: transform .2s ease, box-shadow .2s ease; box-shadow: 0 12px 28px rgba(0,0,0,.24); }
    .button:hover { transform: translateY(-3px); box-shadow: 0 16px 34px rgba(0,0,0,.32); }
    .button svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 2; }
    .hint { margin: 1.5rem 0 0; color: #77758b; font-size: .78rem; }
    @media (max-width: 520px) { .card { border-radius: 24px; } h1 { font-size: 3.3rem; } }
  </style>
</head>
<body>
  <div class="orb orb-one"></div><div class="orb orb-two"></div>
  <main class="shell">
    <section class="card">
      <div class="mark" aria-hidden="true">
        <svg viewBox="0 0 24 24"><path d="M9 18V5l10-2v13M9 18a3 3 0 1 1-3-3 3 3 0 0 1 3 3Zm10-2a3 3 0 1 1-3-3 3 3 0 0 1 3 3Z"/></svg>
      </div>
      <p class="eyebrow">Welcome to Anella</p>
      <h1>Music that feels <span class="gradient">alive.</span></h1>
      <a class="button" href="{{.FrontendURL}}" target="_blank" rel="noopener noreferrer">
        Open Anella <svg viewBox="0 0 24 24"><path d="M7 17 17 7M8 7h9v9"/></svg>
      </a>
      <p class="hint">A brighter way to listen, discover, and remember.</p>
    </section>
  </main>
</body>
</html>`))

func landingPageHTML(frontendURL string) string {
	var page bytes.Buffer
	if err := landingPage.Execute(&page, struct{ FrontendURL string }{FrontendURL: frontendURL}); err != nil {
		return ""
	}

	return page.String()
}
