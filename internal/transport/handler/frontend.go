package handler

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func NewFrontendHandler(serveMode string, diskPath string, cacheTTL time.Duration, embedded fs.FS) http.Handler {
	if serveMode == "disk" && diskPath != "" {
		root, err := filepath.Abs(diskPath)
		if err == nil {
			return &spaHandler{
				root:     http.Dir(root),
				cacheTTL: cacheTTL,
			}
		}
	}

	sub, err := fs.Sub(embedded, "frontend/dist")
	if err != nil {
		panic("embedded frontend not found; build frontend first or set serve_mode=disk")
	}
	return &spaHandler{
		root:     http.FS(sub),
		cacheTTL: cacheTTL,
	}
}

type spaHandler struct {
	root     http.FileSystem
	cacheTTL time.Duration
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := h.root.Open(r.URL.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	} else {
		defer f.Close()
		stat, statErr := f.Stat()
		if statErr == nil && !stat.IsDir() {
			if h.cacheTTL > 0 {
				w.Header().Set("Cache-Control", "public, max-age="+formatSeconds(h.cacheTTL))
			}
			w.Header().Set("Content-Type", contentType(r.URL.Path))
			http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
			return
		}
	}

	spaFile, spaErr := h.root.Open("index.html")
	if spaErr != nil {
		if r.URL.Path == "/" {
			http.NotFound(w, r)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		return
	}
	defer spaFile.Close()
	stat, _ := spaFile.Stat()
	if h.cacheTTL > 0 {
		w.Header().Set("Cache-Control", "public, max-age="+formatSeconds(h.cacheTTL))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", stat.ModTime(), spaFile)
}

var mimeTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "text/javascript; charset=utf-8",
	".json": "application/json",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".ico":  "image/x-icon",
	".woff2": "font/woff2",
}

func contentType(path string) string {
	ext := filepath.Ext(path)
	if mt, ok := mimeTypes[ext]; ok {
		return mt
	}
	return "application/octet-stream"
}

func formatSeconds(d time.Duration) string {
	return time.Duration(d.Seconds()).String()
}
