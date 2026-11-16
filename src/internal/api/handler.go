package api

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"biblioteca-ssd/internal/scanner"
)

type scanHub struct {
	mu      sync.Mutex
	clients map[chan scanner.Progress]struct{}
	latest  scanner.Progress
	active  bool
}

func (h *scanHub) publish(p scanner.Progress) {
	h.mu.Lock()
	h.latest = p
	if p.Done {
		h.active = false
	}
	for c := range h.clients {
		select {
		case c <- p:
		default:
		}
	}
	h.mu.Unlock()
}
func (h *scanHub) subscribe() (chan scanner.Progress, scanner.Progress, func()) {
	c := make(chan scanner.Progress, 8)
	h.mu.Lock()
	h.clients[c] = struct{}{}
	latest := h.latest
	h.mu.Unlock()
	return c, latest, func() { h.mu.Lock(); delete(h.clients, c); close(c); h.mu.Unlock() }
}

// NewHandler serves the health endpoint and embedded SPA (compatibility entry point).
func NewHandler(frontend fs.FS, sessionToken string) http.Handler {
	return NewCatalogHandler(frontend, sessionToken, nil, scanner.Options{})
}

// NewCatalogHandler adds local catalog and scan endpoints to the authenticated SPA.
func NewCatalogHandler(frontend fs.FS, sessionToken string, catalog *sql.DB, options scanner.Options) http.Handler {
	mux := http.NewServeMux()
	hub := &scanHub{clients: map[chan scanner.Progress]struct{}{}}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/scan", func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			http.Error(w, "catálogo indisponível", http.StatusServiceUnavailable)
			return
		}
		hub.mu.Lock()
		if hub.active {
			hub.mu.Unlock()
			http.Error(w, "varredura já em andamento", http.StatusConflict)
			return
		}
		hub.active = true
		hub.latest = scanner.Progress{}
		hub.mu.Unlock()
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
		go func() {
			_, err := scanner.Run(context.Background(), catalog, options, hub.publish)
			if err != nil {
				hub.publish(scanner.Progress{Done: true, Errors: 1, Message: err.Error()})
			}
		}()
	})
	mux.HandleFunc("GET /api/scan/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE indisponível", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		c, latest, done := hub.subscribe()
		defer done()
		sendEvent(w, latest)
		flusher.Flush()
		if latest.Done {
			return
		}
		for {
			select {
			case <-r.Context().Done():
				return
			case p, open := <-c:
				if !open {
					return
				}
				sendEvent(w, p)
				flusher.Flush()
				if p.Done {
					return
				}
			}
		}
	})
	mux.HandleFunc("GET /api/items", func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			http.Error(w, "catálogo indisponível", 503)
			return
		}
		page := queryInt(r, "page", 1)
		size := queryInt(r, "pageSize", 50)
		items, total, err := scanner.List(catalog, r.URL.Query().Get("type"), r.URL.Query().Get("folder"), page, size)
		if err != nil {
			http.Error(w, "filtro inválido", 400)
			return
		}
		writeJSON(w, 200, map[string]any{"items": items, "total": total, "page": page, "pageSize": size})
	})
	mux.HandleFunc("GET /api/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			http.Error(w, "catálogo indisponível", 503)
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.Error(w, "identificador inválido", 400)
			return
		}
		var item scanner.Item
		var ns int64
		var removed int
		err = catalog.QueryRowContext(r.Context(), `SELECT id,path,name,extension,type,size,mtime_ns,partial_hash,removed FROM items WHERE id=?`, id).Scan(&item.ID, &item.Path, &item.Name, &item.Extension, &item.Type, &item.Size, &ns, &item.PartialHash, &removed)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, "falha ao consultar item", 500)
			return
		}
		item.ModifiedAt = unixNano(ns)
		item.Removed = removed != 0
		writeJSON(w, 200, item)
	})
	mux.HandleFunc("GET /api/folders", func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			http.Error(w, "catálogo indisponível", 503)
			return
		}
		rows, err := catalog.QueryContext(r.Context(), `SELECT path FROM items WHERE removed=0 ORDER BY path COLLATE NOCASE`)
		if err != nil {
			http.Error(w, "falha ao listar pastas", 500)
			return
		}
		defer rows.Close()
		set := map[string]bool{"": true}
		for rows.Next() {
			var p string
			if rows.Scan(&p) != nil {
				continue
			}
			for d := path.Dir(p); d != "." && d != "/"; d = path.Dir(d) {
				set[d] = true
			}
		}
		folders := make([]string, 0, len(set))
		for f := range set {
			folders = append(folders, f)
		}
		sort.Strings(folders)
		writeJSON(w, 200, map[string]any{"folders": folders})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "." || name == "" {
			name = "index.html"
		}
		file, err := frontend.Open(name)
		if err != nil {
			name = "index.html"
			file, err = frontend.Open(name)
		}
		if err != nil {
			http.Error(w, "frontend indisponível", 500)
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
			http.Error(w, "não foi possível ler o frontend", 500)
			return
		}
		_, _ = w.Write(contents)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			mux.ServeHTTP(w, r)
			return
		}
		if r.URL.Query().Get("session") == sessionToken && sessionToken != "" {
			http.SetCookie(w, &http.Cookie{Name: "biblioteca_session", Value: sessionToken, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
			clean := *r.URL
			q := clean.Query()
			q.Del("session")
			clean.RawQuery = q.Encode()
			http.Redirect(w, r, clean.String(), http.StatusSeeOther)
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func queryInt(r *http.Request, key string, fallback int) int {
	v, e := strconv.Atoi(r.URL.Query().Get(key))
	if e != nil || v < 1 {
		return fallback
	}
	return v
}
func sendEvent(w io.Writer, p scanner.Progress) {
	data, _ := json.Marshal(p)
	_, _ = fmt.Fprintf(w, "event: progress\ndata: %s\n\n", data)
}
func unixNano(v int64) time.Time { return time.Unix(0, v) }
