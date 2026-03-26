package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var dist embed.FS

func NewHandler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return &spaHandler{sub: sub, files: http.FileServer(http.FS(sub))}
}

type spaHandler struct {
	sub   fs.FS
	files http.Handler
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path != "" {
		f, err := h.sub.Open(path)
		if err == nil {
			stat, serr := f.Stat()
			f.Close()
			if serr == nil && !stat.IsDir() {
				h.files.ServeHTTP(w, r)
				return
			}
		}
	}

	http.ServeFileFS(w, r, h.sub, "index.html")
}
