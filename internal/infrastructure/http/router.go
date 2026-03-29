package http

import (
	"github.com/go-chi/chi/v5"
)

// NewMux creates a chi router with standard middleware applied.
// Route registration is done by the caller (typically handler.Router).
func NewMux() *chi.Mux {
	return chi.NewRouter()
}
