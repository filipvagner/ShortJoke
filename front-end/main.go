package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sync"
)

// URLStore holds our links in memory.
// In a "real" app, you'd use a database like SQLite here.
type URLStore struct {
	mu   sync.Mutex
	urls map[string]string
}

var store = URLStore{urls: make(map[string]string)}
var tmpl = template.Must(template.New("index").Parse(htmlTemplate))

// Generate a random short string
func generateID() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	// 1. Handle the Home Page and Shortening logic
	http.HandleFunc("/", homeHandler)

	// 2. Handle the Redirection logic (e.g., localhost:8080/abc123)
	http.HandleFunc("/r/", redirectHandler)

	fmt.Println("🚀 Server starting at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		longURL := r.FormValue("longUrl")
		if longURL == "" {
			http.Error(w, "URL cannot be empty", http.StatusBadRequest)
			return
		}

		parsed, err := url.Parse(longURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			http.Error(w, "Unsupported URL scheme", http.StatusBadRequest)
			return
		}

		id := generateID()
		store.mu.Lock()
		store.urls[id] = longURL
		store.mu.Unlock()

		shortURL := fmt.Sprintf("http://localhost:8080/r/%s", id)
		tmpl.Execute(w, map[string]string{"ShortURL": shortURL})
		return
	}
	tmpl.Execute(w, nil)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/r/"):]
	store.mu.Lock()
	longURL, exists := store.urls[id]
	store.mu.Unlock()

	if !exists {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

// Simple HTML Template with basic CSS
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Go Shortener</title>
	<style>
		body { font-family: sans-serif; display: flex; justify-content: center; padding: 50px; background: #f4f4f9; }
		.card { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); width: 400px; }
		input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
		button { background: #007bff; color: white; border: none; padding: 10px; width: 100%; cursor: pointer; border-radius: 4px; }
		.result { margin-top: 20px; padding: 10px; background: #e2e3e5; border-radius: 4px; word-break: break-all; }
	</style>
</head>
<body>
	<div class="card">
		<h2>URL Shortener</h2>
		<form method="POST">
			<input type="url" name="longUrl" placeholder="https://example.com" required>
			<button type="submit">Shorten</button>
		</form>
		{{if .ShortURL}}
			<div class="result">
				<strong>Your short link:</strong><br>
				<a href="{{.ShortURL}}" target="_blank">{{.ShortURL}}</a>
			</div>
		{{end}}
	</div>
</body>
</html>
`
