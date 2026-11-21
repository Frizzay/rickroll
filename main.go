// Rickroll web app: every URL on this server shows only a harmless title in
// chat link previews. Real browsers end up at "Never Gonna Give You Up".
//
// How the preview stays clean:
//   - The shared URL never answers with an HTTP redirect. Preview bots follow
//     redirects and would otherwise show YouTube's metadata.
//   - Preview bots (detected by user agent) get a page without any script.
//     This matters for iMessage, whose preview executes JavaScript.
//   - Browsers get the same page plus a small script that redirects to /r.
//     Only /r redirects to YouTube with a 302, so the HTML never hints at
//     YouTube.
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const target = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

type config struct {
	Title       string
	Description string
	Image       string // absolute URL, optional
	SiteName    string // optional
}

// botMarkers are user agent fragments that identify link preview services and
// crawlers. Matching is done in lowercase.
//
// The list is deliberately narrow. A human misclassified as a bot would be
// stuck on the loading page. An undetected bot is usually harmless, because
// almost no preview bot executes JavaScript. Names of apps with their own
// in-app browser (LINE, Snapchat, LinkedIn, …) are therefore left out. Their
// preview bots are called "…bot" or "…preview" anyway.
var botMarkers = []string{
	// Twitterbot, Discordbot, TelegramBot, Slackbot, LinkedInBot, Googlebot, …
	"bot", "crawler", "spider",
	// SkypeUriPreview (Teams, Skype), BingPreview, …
	"preview",
	// Facebook, Messenger, Instagram and iMessage
	"facebookexternalhit", "facebot",
	// WhatsApp; Signal uses this user agent as well
	"whatsapp",
	// Image proxies, embed services, Fediverse servers
	"slack-imgproxy", "embedly", "iframely", "mastodon", "pleroma", "misskey",
	// Command line tools and headless browsers
	"curl/", "wget/", "python-", "go-http-client", "okhttp", "headlesschrome",
}

func isBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	if ua == "" {
		return true
	}
	for _, m := range botMarkers {
		if strings.Contains(ua, m) {
			return true
		}
	}
	return false
}

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex, nofollow">
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<meta property="og:type" content="website">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
{{- with .SiteName}}
<meta property="og:site_name" content="{{.}}">
{{- end}}
{{- with .Image}}
<meta property="og:image" content="{{.}}">
<meta name="twitter:card" content="summary_large_image">
{{- else}}
<meta name="twitter:card" content="summary">
{{- end}}
<meta name="twitter:title" content="{{.Title}}">
<meta name="twitter:description" content="{{.Description}}">
<style>
  html, body { height: 100%; margin: 0; }
  body {
    display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px;
    font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
    background: #f5f5f7; color: #333;
  }
  .spinner {
    width: 36px; height: 36px; border-radius: 50%;
    border: 4px solid #ddd; border-top-color: #555;
    animation: spin .8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  a { color: #888; font-size: 14px; }
</style>
</head>
<body>
<div class="spinner"></div>
<p>Loading …</p>
<a href="/r">Continue</a>
{{- if .Redirect}}
<script>location.replace("/r");</script>
{{- end}}
</body>
</html>
`))

type pageData struct {
	config
	Redirect bool
}

func newHandler(cfg config) http.Handler {
	mux := http.NewServeMux()

	// Only the browser script (or the Continue link) requests this path.
	mux.HandleFunc("GET /r", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.Redirect(w, r, target, http.StatusFound)
	})

	// Otherwise the browser would load the decoy page here as well.
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Every other path serves the decoy page, so links like
	// https://example.com/photos/summer-2026 work.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "no-store")
		h.Add("Vary", "User-Agent") // bots and browsers get different content
		data := pageData{config: cfg, Redirect: !isBot(r.UserAgent())}
		if err := page.Execute(w, data); err != nil {
			log.Printf("template error: %v", err)
		}
	})

	return mux
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	addr := flag.String("addr", ":"+envOr("PORT", "8080"), "address the server listens on")
	var cfg config
	flag.StringVar(&cfg.Title, "title", envOr("PREVIEW_TITLE", "Weekend photos 📸"), "title shown in the link preview")
	flag.StringVar(&cfg.Description, "description", envOr("PREVIEW_DESCRIPTION", "Here are the pictures I promised you."), "description shown in the link preview")
	flag.StringVar(&cfg.Image, "image", envOr("PREVIEW_IMAGE", ""), "absolute URL of a preview image (optional)")
	flag.StringVar(&cfg.SiteName, "site-name", envOr("PREVIEW_SITE_NAME", ""), "site name shown in the link preview (optional)")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           newHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("server listening on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}
