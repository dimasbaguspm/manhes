package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"manga-engine/internal/domain"
	"manga-engine/internal/handler"
	"manga-engine/pkg/httputil"
)

// httpHandlers wraps handler.Services with HTTP-specific methods.
type httpHandlers struct {
	*handler.Handlers
}

func (h *httpHandlers) internalError(w http.ResponseWriter, r *http.Request, op string, err error) {
	h.Log.Error(op,
		"request_id", r.Context().Value("request_id"),
		"err", err,
	)
	httputil.InternalError(w, err)
}

// listManga HTTP handler: GET /api/v1/manga
//
// @Summary     List manga
// @Description Returns a paginated list of manga. Supports filtering by id, q, genre, author, state and sorting.
// @Tags        manga
// @Produce     json
// @Param       id          query  []string false "Filter by dictionary_id (comma-separated or repeated)"
// @Param       q           query  string  false "Search by title or description"
// @Param       genre       query  []string false "Filter by genre (comma-separated or repeated)"
// @Param       author      query  []string false "Filter by author (comma-separated or repeated)"
// @Param       state       query  []string false "Filter by state (unavailable|fetching|available)"
// @Param       sortBy      query  string  false "Sort field: title | updatedAt | createdAt (default: title)"
// @Param       sortOrder   query  string  false "Sort order: asc | desc (default: asc)"
// @Param       page        query  int     false "Page number (default 1)"
// @Param       pageSize    query  int     false "Items per page (default 20, max 100)"
// @Success     200  {object}  domain.MangaListResponse
// @Failure     500  {object}  httputil.ErrorResponse
// @Router      /manga [get]
func (h *httpHandlers) listManga(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := httputil.IntQueryParam(q.Get("page"), 1)
	pageSize := httputil.IntQueryParam(q.Get("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	ids := parseStringArray(q["id"])
	qVal := q.Get("q")
	genres := parseStringArray(q["genre"])
	authors := parseStringArray(q["author"])
	states := parseStringArray(q["state"])
	sortBy := q.Get("sortBy")
	if sortBy == "" {
		sortBy = "title"
	}
	sortOrder := q.Get("sortOrder")
	if sortOrder == "" {
		sortOrder = "asc"
	}

	filter := domain.MangaFilter{
		IDs:       ids,
		Q:         qVal,
		Genres:    genres,
		Authors:   authors,
		States:    states,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		PageSize:  pageSize,
	}

	result, items, err := h.Handlers.ListManga(r.Context(), filter)
	if err != nil {
		h.internalError(w, r, "list manga", err)
		return
	}

	pageTotal := result.Total / result.PageSize
	if result.Total%result.PageSize != 0 {
		pageTotal++
	}

	httputil.WriteJSON(w, http.StatusOK, domain.MangaListResponse{
		Pagination: domain.Pagination{
			PageNumber: result.Page,
			PageSize:   result.PageSize,
			PageTotal:  pageTotal,
			ItemCount:  result.Total,
		},
		Items: items,
	})
}

// getManga HTTP handler: GET /api/v1/manga/{mangaId}
//
// @Summary     Get manga detail
// @Description mangaId is the manga UUID. Unavailable/fetching manga return partial detail.
// @Tags        manga
// @Produce     json
// @Param       mangaId  path      string  true  "Manga UUID"
// @Success     200      {object}  domain.MangaDetailResponse
// @Failure     404      {object}  httputil.ErrorResponse
// @Failure     500      {object}  httputil.ErrorResponse
// @Router      /manga/{mangaId} [get]
func (h *httpHandlers) getManga(w http.ResponseWriter, r *http.Request) {
	mangaID := chi.URLParam(r, "mangaId")

	detail, found, err := h.Handlers.GetManga(r.Context(), mangaID)
	if err != nil {
		h.internalError(w, r, "get manga", err)
		return
	}
	if !found {
		httputil.NotFound(w, "manga not found", nil)
		return
	}

	resp := domain.MangaDetailResponse{
		ID:           detail.ID,
		DictionaryID: detail.DictionaryID,
		Title:        detail.Title,
		State:        string(detail.State),
		Description:  detail.Description,
		Status:       detail.Status,
		Authors:      detail.Authors,
		Genres:       detail.Genres,
		CoverURL:     detail.CoverURL,
		UpdatedAt:    detail.UpdatedAt,
		CreatedAt:    detail.CreatedAt,
	}
	for _, l := range detail.Languages {
		resp.Languages = append(resp.Languages, domain.MangaLangResponse{
			Lang:              l.Language,
			TotalChapters:     l.Total,
			AvailableChapters: l.Available,
			LatestUpdate:      l.LatestUpdate,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// getChapters HTTP handler: GET /api/v1/manga/{mangaId}/{lang}
//
// @Summary     List uploaded chapters for a language
// @Description Returns all uploaded chapters for a specific language. Requires manga to be in available state.
// @Tags        manga
// @Produce     json
// @Param       mangaId  path   string  true  "Manga UUID"
// @Param       lang     path   string  true  "Language code (e.g. en, fr)"
// @Success     200      {object}  domain.ChapterListResponse
// @Failure     404      {object}  httputil.ErrorResponse
// @Failure     500      {object}  httputil.ErrorResponse
// @Router      /manga/{mangaId}/{lang} [get]
func (h *httpHandlers) getChapters(w http.ResponseWriter, r *http.Request) {
	mangaID := chi.URLParam(r, "mangaId")
	lang := chi.URLParam(r, "lang")

	chapters, found, err := h.Handlers.GetChaptersByLang(r.Context(), mangaID, lang)
	if err != nil {
		h.internalError(w, r, "get chapters by lang", err)
		return
	}
	if !found {
		httputil.NotFound(w, "manga not found or content not yet available", nil)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, domain.ChapterListResponse{
		ID:       mangaID,
		Lang:     lang,
		Chapters: chapters,
	})
}

// readChapter HTTP handler: GET /api/v1/read/{chapterId}
//
// @Summary     Get chapter pages
// @Description Returns page URLs for the requested chapter along with prev/next navigation links.
// @Tags        manga
// @Produce     json
// @Param       chapterId  path   string  true  "Chapter UUID"
// @Success     200      {object}  domain.ChapterReadResponse
// @Failure     400      {object}  httputil.ErrorResponse
// @Failure     404      {object}  httputil.ErrorResponse
// @Failure     500      {object}  httputil.ErrorResponse
// @Router      /read/{chapterId} [get]
func (h *httpHandlers) readChapter(w http.ResponseWriter, r *http.Request) {
	chapterID := chi.URLParam(r, "chapterId")

	result, found, err := h.Handlers.ReadChapter(r.Context(), chapterID)
	if err != nil {
		h.internalError(w, r, "read chapter", err)
		return
	}
	if !found {
		httputil.NotFound(w, "chapter not found or content not yet available", nil)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, domain.ChapterReadResponse{
		MangaID:     result.MangaID,
		ChapterID:   chapterID,
		Pages:       result.Pages,
		PrevChapter: result.PrevChapter,
		NextChapter: result.NextChapter,
	})
}

// refreshDictionary HTTP handler: POST /api/v1/dictionary/refresh
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
func (h *httpHandlers) refreshDictionary(w http.ResponseWriter, r *http.Request) {
	var req domain.DictionaryRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DictionaryID == "" {
		httputil.BadRequest(w, "dictionaryId is required", nil)
		return
	}

	if err := h.Bus.Publish(r.Context(), h.Cfg.Bus.DictionaryRefreshed, domain.DictionaryRefreshed{
		DictionaryID: req.DictionaryID,
	}); err != nil {
		h.internalError(w, r, "refresh dictionary", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// searchDictionary HTTP handler: GET /api/v1/dictionary?q=...
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
func (h *httpHandlers) searchDictionary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.BadRequest(w, "q query param is required", nil)
		return
	}
	entries, err := h.Handlers.Search(r.Context(), q)
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

// parseStringArray parses a query parameter that may be repeated or comma-separated.
func parseStringArray(vals []string) []string {
	var result []string
	for _, v := range vals {
		if v == "" {
			continue
		}
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
	}
	return result
}

// upsertTracker HTTP handler: PUT /api/v1/tracker
//
// @Summary     Upsert a tracker entry
// @Description Creates or updates a reading tracker entry. Metadata is fully replaced on each call.
// @Tags        tracker
// @Accept      json
// @Produce     json
// @Param       body  body  domain.UpsertTrackerRequest  true  "Tracker data"
// @Success     200  {object}  domain.TrackerResponse
// @Failure     400  {object}  httputil.ErrorResponse
// @Failure     500  {object}  httputil.ErrorResponse
// @Router      /tracker [put]
func (h *httpHandlers) upsertTracker(w http.ResponseWriter, r *http.Request) {
	var req domain.UpsertTrackerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body", err)
		return
	}
	if req.MangaID == "" || req.ChapterID == "" {
		httputil.BadRequest(w, "manga_id and chapter_id are required", nil)
		return
	}

	tracker := domain.Tracker{
		ID:        req.ID,
		MangaID:   req.MangaID,
		ChapterID: req.ChapterID,
		IsRead:    req.IsRead,
		Metadata:  req.Metadata,
	}

	if err := h.Repo.UpsertTracker(r.Context(), tracker); err != nil {
		h.internalError(w, r, "upsert tracker", err)
		return
	}

	// Fetch the updated tracker to return fresh timestamps
	result, found, err := h.Repo.GetTracker(r.Context(), req.MangaID, req.ChapterID)
	if err != nil {
		h.internalError(w, r, "upsert tracker", err)
		return
	}
	if !found {
		httputil.InternalError(w, nil)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, domain.TrackerResponse{
		ID:        result.ID,
		MangaID:   result.MangaID,
		ChapterID: result.ChapterID,
		IsRead:    result.IsRead,
		Metadata:  string(result.Metadata),
		UpdatedAt: result.UpdatedAt,
		CreatedAt: result.CreatedAt,
	})
}

// getTrackersByManga HTTP handler: GET /api/v1/tracker/{mangaId}
//
// @Summary     List trackers for a manga
// @Description Returns all tracker entries for a specific manga.
// @Tags        tracker
// @Produce     json
// @Param       mangaId  path  string  true  "Manga UUID"
// @Success     200  {array}  domain.TrackerResponse
// @Failure     500  {object}  httputil.ErrorResponse
// @Router      /tracker/{mangaId} [get]
func (h *httpHandlers) getTrackersByManga(w http.ResponseWriter, r *http.Request) {
	mangaID := chi.URLParam(r, "mangaId")

	results, err := h.Repo.GetTrackersByManga(r.Context(), mangaID)
	if err != nil {
		h.internalError(w, r, "get trackers by manga", err)
		return
	}

	trackers := make([]domain.TrackerResponse, 0, len(results))
	for _, t := range results {
		trackers = append(trackers, domain.TrackerResponse{
			ID:        t.ID,
			MangaID:   t.MangaID,
			ChapterID: t.ChapterID,
			IsRead:    t.IsRead,
			Metadata:  string(t.Metadata),
			UpdatedAt: t.UpdatedAt,
			CreatedAt: t.CreatedAt,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, trackers)
}
