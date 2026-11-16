package api

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewHandler serves the health endpoint and the embedded single-page frontend.
func NewHandler(frontend fs.FS, sessionToken string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "." || name == "" {
			name = "index.html"
		}
		file, err := frontend.Open(name)
		if err != nil {
			name = "index.html" // SPA fallback for client-side routes.
			file, err = frontend.Open(name)
		}
		if err != nil {
			http.Error(w, "frontend indisponível", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if name == "index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else if strings.HasSuffix(name, ".js") {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		} else if strings.HasSuffix(name, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		contents, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "não foi possível ler o frontend", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(contents)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			mux.ServeHTTP(w, r)
			return
		}
		if r.URL.Query().Get("session") == sessionToken && sessionToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name: "biblioteca_session", Value: sessionToken, Path: "/",
				HttpOnly: true, SameSite: http.SameSiteStrictMode,
			})
			cleanURL := *r.URL
			query := cleanURL.Query()
			query.Del("session")
			cleanURL.RawQuery = query.Encode()
			http.Redirect(w, r, cleanURL.String(), http.StatusSeeOther)
			return
		}
		cookie, err := r.Cookie("biblioteca_session")
		if err != nil || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(sessionToken)) != 1 {
			http.Error(w, "sessão local necessária", http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
