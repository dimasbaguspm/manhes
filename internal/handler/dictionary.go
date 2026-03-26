package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"manga-engine/internal/domain"
	"manga-engine/pkg/httputil"
)

func dictToResponse(e domain.DictionaryEntry) domain.DictionaryResponse {
	r := domain.DictionaryResponse{
		ID:          e.ID,
		Slug:        e.Slug,
		Title:       e.Title,
		State:       string(e.State),
		CoverURL:    e.CoverURL,
		Sources:     e.Sources,
		BestSource:  e.BestSource,
		RefreshedAt: e.RefreshedAt,
	}
	if len(e.SourceStats) > 0 {
		r.SourceStats = e.SourceStats
		chapters := make(map[string]int)
		for _, v := range e.SourceStats {
			if v.Err == "" {
				for lang, count := range v.ChaptersByLang {
					if count > chapters[lang] {
						chapters[lang] = count
					}
				}
			}
		}
		if len(chapters) > 0 {
			r.ChaptersByLang = chapters
		}
	}
	return r
}

// RefreshDictionary handles POST /api/v1/dictionary/{dictionaryId}
//
// @Summary     Refresh a dictionary entry
// @Description Re-fetches source stats from all known scrapers and searches
//
//	3rd-party sources for any new entries matching the title.
//	Safe to call repeatedly — already-known sources are preserved.
//
// @Tags        dictionary
// @Produce     json
// @Param       dictionaryId  path  string  true  "Dictionary entry ID"
// @Success     200  {object}  domain.DictionaryResponse
// @Failure     404  {object}  httputil.ErrorResponse
// @Failure     500  {object}  httputil.ErrorResponse
// @Router      /dictionary/{dictionaryId} [post]
func (h *Handlers) RefreshDictionary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "dictionaryId")
	entry, err := h.dictionary.Refresh(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputil.NotFound(w, "dictionary entry not found", nil)
			return
		}
		h.internalError(w, r, "refresh dictionary", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dictToResponse(entry))
}

// SearchDictionary handles GET /api/v1/dictionary?q=...
//
// @Summary     Search all scrapers and auto-upsert results into dictionary
// @Description Results are permanently stored in the dictionary. Re-searching a known
//
//	title updates its source IDs without creating duplicates.
//
// @Tags        dictionary
// @Produce     json
// @Param       q  query  string  true  "Search query"
// @Success     200  {array}   domain.DictionaryResponse
// @Failure     400  {object}  httputil.ErrorResponse
// @Failure     500  {object}  httputil.ErrorResponse
// @Router      /dictionary [get]
func (h *Handlers) SearchDictionary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.BadRequest(w, "q query param is required", nil)
		return
	}
	entries, err := h.dictionary.Search(r.Context(), q)
	if err != nil {
		h.internalError(w, r, "search dictionary", err)
		return
	}
	result := make([]domain.DictionaryResponse, 0, len(entries))
	for _, e := range entries {
		result = append(result, dictToResponse(e))
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
