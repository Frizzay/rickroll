package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testCfg = config{
	Title:       "Weekend photos",
	Description: "Here are the pictures.",
	Image:       "https://example.com/preview.jpg",
}

func get(t *testing.T, path, userAgent string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", userAgent)
	rec := httptest.NewRecorder()
	newHandler(testCfg).ServeHTTP(rec, req)
	return rec.Result()
}

func body(t *testing.T, res *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPreviewBotsSeeOnlyDecoy(t *testing.T) {
	bots := map[string]string{
		"WhatsApp": "WhatsApp/2.23.20.0 A",
		"Telegram": "TelegramBot (like TwitterBot)",
		"Discord":  "Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)",
		"Slack":    "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)",
		"iMessage": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/601.2.4 (KHTML, like Gecko) Version/9.0.1 Safari/601.2.4 facebookexternalhit/1.1 Facebot Twitterbot/1.0",
		"empty":    "",
	}
	for name, ua := range bots {
		t.Run(name, func(t *testing.T) {
			res := get(t, "/", ua)
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200 (no redirect for bots)", res.StatusCode)
			}
			b := body(t, res)
			if !strings.Contains(b, `property="og:title" content="Weekend photos"`) {
				t.Errorf("og:title missing:\n%s", b)
			}
			if strings.Contains(b, "<script") {
				t.Errorf("bot gets a script:\n%s", b)
			}
			if strings.Contains(strings.ToLower(b), "youtu") {
				t.Errorf("page reveals YouTube:\n%s", b)
			}
		})
	}
}

func TestBrowserGetsRedirectScript(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	for _, path := range []string{"/", "/photos/summer-2026"} {
		res := get(t, path, ua)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", path, res.StatusCode)
		}
		b := body(t, res)
		if !strings.Contains(b, `location.replace("/r")`) {
			t.Errorf("%s: redirect script missing:\n%s", path, b)
		}
		if strings.Contains(strings.ToLower(b), "youtu") {
			t.Errorf("%s: page reveals YouTube:\n%s", path, b)
		}
	}
}

func TestRedirectEndpoint(t *testing.T) {
	res := get(t, "/r", "Mozilla/5.0")
	if res.StatusCode != http.StatusFound {
		t.Fatalf("status = %d, want 302", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != target {
		t.Errorf("Location = %q, want %q", loc, target)
	}
}
