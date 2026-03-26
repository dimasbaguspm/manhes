package handler

import (
	"encoding/json"
	"net/http"

	"manga-engine/pkg/httputil"
)

type watchlistRequest struct {
	DictionaryID string `json:"dictionaryId"`
}

// AddWatchlist handles POST /api/v1/watchlist
//
// @Summary     Register manga for ingestion
// @Description Accepts a dictionary entry ID, saves the watchlist entry, sets dictionary state to fetching, and publishes an IngestRequested event. Returns 202 immediately.
// @Tags        watchlist
// @Accept      json
// @Produce     json
// @Param       body  body  watchlistRequest  true  "Watchlist entry"
// @Success     202   {object}  map[string]string
// @Failure     400   {object}  httputil.ErrorResponse
// @Failure     500   {object}  httputil.ErrorResponse
// @Router      /watchlist [post]
func (h *Handlers) AddWatchlist(w http.ResponseWriter, r *http.Request) {
	var req watchlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body", err)
		return
	}
	if req.DictionaryID == "" {
		httputil.BadRequest(w, "dictionaryId is required", nil)
		return
	}

	slug, err := h.watchlist.AddByDictionaryID(r.Context(), req.DictionaryID)
	if err != nil {
		h.internalError(w, r, "add watchlist", err)
		return
	}
	httputil.WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":       "accepted",
		"slug":         slug,
		"dictionaryId": req.DictionaryID,
	})
}
