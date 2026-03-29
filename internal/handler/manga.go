package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"manga-engine/internal/domain"
	"manga-engine/pkg/httputil"
)

// ListManga handles GET /api/v1/manga
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
func (h *Handlers) ListManga(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := httputil.IntQueryParam(q.Get("page"), 1)
	pageSize := httputil.IntQueryParam(q.Get("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse repeated query params
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

	result, err := h.manga.ListManga(r.Context(), filter)
	if err != nil {
		h.internalError(w, r, "list manga", err)
		return
	}

	items := make([]domain.MangaSummary, 0, len(result.Items))
	for _, m := range result.Items {
		// Get per-language chapter stats for this manga.
		languages, err := h.manga.GetMangaLanguages(r.Context(), m.ID, m.DictionaryID)
		if err != nil {
			h.internalError(w, r, "get manga languages", err)
			return
		}
		items = append(items, domain.MangaSummary{
			ID:             m.ID,
			DictionaryID:   m.DictionaryID,
			Title:          m.Title,
			Description:    m.Description,
			Status:         m.Status,
			CoverURL:       m.CoverURL,
			State:          string(m.State),
			Authors:        m.Authors,
			Genres:         m.Genres,
			Languages:      languages,
			UpdatedAt:      m.UpdatedAt,
			CreatedAt:      m.CreatedAt,
		})
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

// GetManga handles GET /api/v1/manga/{mangaId}
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
func (h *Handlers) GetManga(w http.ResponseWriter, r *http.Request) {
	mangaID := chi.URLParam(r, "mangaId")

	detail, found, err := h.manga.GetManga(r.Context(), mangaID)
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

// GetChaptersByLang handles GET /api/v1/manga/{mangaId}/{lang}
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
func (h *Handlers) GetChaptersByLang(w http.ResponseWriter, r *http.Request) {
	mangaID := chi.URLParam(r, "mangaId")
	lang := chi.URLParam(r, "lang")

	chapters, found, err := h.manga.GetChaptersByLang(r.Context(), mangaID, lang)
	if err != nil {
		h.internalError(w, r, "get chapters by lang", err)
		return
	}
	if !found {
		httputil.NotFound(w, "manga not found or content not yet available", nil)
		return
	}

	items := make([]domain.ChapterItem, 0, len(chapters))
	for _, ch := range chapters {
		items = append(items, domain.ChapterItem{
			ID:        ch.ID,
			Order:     ch.Order,
			Name:      ch.Name,
			UpdatedAt: ch.UpdatedAt,
			PageCount: ch.PageCount,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, domain.ChapterListResponse{
		ID:       mangaID,
		Lang:     lang,
		Chapters: items,
	})
}

// ReadChapter handles GET /api/v1/read/{chapterId}
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
func (h *Handlers) ReadChapter(w http.ResponseWriter, r *http.Request) {
	chapterID := chi.URLParam(r, "chapterId")

	result, found, err := h.manga.ReadChapter(r.Context(), chapterID)
	if err != nil {
		h.internalError(w, r, "read chapter", err)
		return
	}
	if !found {
		httputil.NotFound(w, "chapter not found or content not yet available", nil)
		return
	}

	resp := domain.ChapterReadResponse{
		MangaID:   result.MangaID,
		ChapterID:  chapterID,
		Pages:      result.Pages,
		PrevChapter: result.PrevChapter,
		NextChapter: result.NextChapter,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// parseStringArray parses a query parameter that may be repeated or comma-separated.
func parseStringArray(vals []string) []string {
	var result []string
	for _, v := range vals {
		if v == "" {
			continue
		}
		// Handle comma-separated values
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
