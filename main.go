package main

import (
	"encoding/base64"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"text/template"
)

type URLStore struct {
	urls  map[string]string
	mutex sync.RWMutex
}

func NewURLStore() *URLStore {
	return &URLStore{
		urls: make(map[string]string),
	}
}

func generateShortURL() string {
	b := make([]byte, 4)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:6]
}

func main() {
	store := NewURLStore()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			store.mutex.RLock()
			if target, ok := store.urls[r.URL.Path[1:]]; ok {
				store.mutex.RUnlock()
				http.Redirect(w, r, target, http.StatusFound)
				return
			}
			store.mutex.RUnlock()
			http.NotFound(w, r)
			return
		}

		tmpl := template.Must(template.ParseFiles("template/index.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		originalURL := r.FormValue("url")
		if originalURL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		shortURL := generateShortURL()

		store.mutex.Lock()
		store.urls[shortURL] = originalURL
		store.mutex.Unlock()

		w.Write([]byte(shortURL))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
