package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"manga-engine/config"
	"manga-engine/internal/ui"
)

func NewRouter(h *Handlers, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(cors)
	r.Use(requestID)
	r.Use(structuredLogger(h.log))
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/manga", h.ListManga)
		r.Get("/manga/{mangaId}", h.GetManga)
		r.Get("/manga/{mangaId}/{lang}", h.GetChaptersByLang)
		r.Get("/manga/{mangaId}/{lang}/read", h.ReadChapter)

		r.Get("/dictionary", h.SearchDictionary)
		r.Post("/dictionary/refresh", h.RefreshDictionary)
	})

	// SPA is only served in prod — in dev the Vite server handles the frontend.
	if cfg.Env != "dev" {
		r.Handle("/*", ui.NewHandler())
	}

	return r
}
