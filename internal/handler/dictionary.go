package handler

import (
	"encoding/json"
	"net/http"

	"manga-engine/internal/domain"
	"manga-engine/pkg/httputil"
)

func dictToResponse(e domain.DictionaryEntry) domain.DictionaryResponse {
	r := domain.DictionaryResponse{
		ID:         e.ID,
		Slug:       e.Slug,
		Title:      e.Title,
		CoverURL:   e.CoverURL,
		Sources:    e.Sources,
		BestSource: e.BestSource,
		UpdatedAt:  e.UpdatedAt,
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

// RefreshDictionary handles POST /api/v1/dictionary/refresh
//
// @Summary     Refresh a dictionary entry
// @Description Re-fetches source stats from all known scrapers.
//
//	The actual refresh runs in the background; this endpoint returns 202 immediately.
//
// @Tags        dictionary
// @Accept      json
// @Produce     json
// @Param       dictionaryId  body  domain.DictionaryRefreshRequest  true  "Dictionary entry ID"
// @Success     202
// @Failure     400  {object}  httputil.ErrorResponse
// @Failure     500  {object}  httputil.ErrorResponse
// @Router      /dictionary/refresh [post]
func (h *Handlers) RefreshDictionary(w http.ResponseWriter, r *http.Request) {
	var req domain.DictionaryRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DictionaryID == "" {
		httputil.BadRequest(w, "dictionaryId is required", nil)
		return
	}

	if err := h.bus.Publish(r.Context(), h.cfg.Bus.DictionaryRefreshed, domain.DictionaryRefreshed{
		DictionaryID: req.DictionaryID,
	}); err != nil {
		h.internalError(w, r, "refresh dictionary", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
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
