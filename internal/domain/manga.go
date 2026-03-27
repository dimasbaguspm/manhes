package domain

import (
	"context"
	"time"
)

// Manga is the core aggregate representing a manga title.
type Manga struct {
	Slug           string
	Title          string
	Description    string
	Status         string            // ongoing | completed | hiatus
	Authors        []string
	Genres         []string
	Languages      []string
	CoverURL       string
	Sources        map[string]string // source_name → source_id
	ScrapedAt      time.Time
	UpdatedAt      *time.Time
	DictionaryID   string
	State          MangaState
	ChaptersByLang map[string]int // lang → available chapter count
}

// MangaFilter holds optional filters and pagination for ListManga.
type MangaFilter struct {
	Title           string
	Status          string
	State           string // "unavailable" | "fetching" | "available" | "" (all)
	SortBy          string // "title" (default) | "last_update"
	Page            int    // 1-based
	PageSize        int
	HideUnavailable bool // exclude entries with state = "unavailable"
}

// MangaPage is a paginated slice of Manga.
type MangaPage struct {
	Items    []Manga
	Total    int
	Page     int
	PageSize int
}

// ChapterRef is a lightweight reference to a chapter pending S3 upload.
type ChapterRef struct {
	DictionaryID string // used as S3 path prefix; falls back to Slug if empty
	Slug         string
	Language     string
	ChapterNum   string
}

// MangaDetail includes full manga info with languages and chapters.
type MangaDetail struct {
	Manga
	Languages []MangaLang
	Chapters  []MangaChapter
}

// MangaLang holds per-language availability stats.
type MangaLang struct {
	Language     string
	Available    int // total chapters known from source
	Fetched      int // chapters downloaded to disk
	Uploaded     int // chapters uploaded to S3 (readable)
	LatestUpdate *time.Time
}

// MangaChapter holds chapter catalog metadata.
type MangaChapter struct {
	Slug       string
	Language   string
	ChapterNum string
	PageCount  int
	Uploaded   bool
	UploadedAt *time.Time
}

// ChapterRead holds the result of reading a chapter: page URLs and prev/next navigation.
type ChapterRead struct {
	Pages       []string
	PrevChapter *string
	NextChapter *string
}

// CatalogQuerier is the read-only manga catalog port used by the HTTP handler.
type CatalogQuerier interface {
	ListManga(ctx context.Context, filter MangaFilter) (MangaPage, error)
	GetManga(ctx context.Context, dictionaryID string) (MangaDetail, bool, error)
	GetChaptersByLang(ctx context.Context, dictionaryID, lang string) ([]MangaChapter, bool, error)
	ReadChapter(ctx context.Context, dictionaryID, lang string, num string) (ChapterRead, bool, error)
}

// Pagination is a reusable pagination envelope for list responses.
type Pagination struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	PageTotal  int `json:"pageTotal"`
	ItemCount  int `json:"itemCount"`
}

// MangaSummary is the list-view representation of a manga entry.
type MangaSummary struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Status         string         `json:"status"`
	CoverURL       string         `json:"cover_url"`
	State          string         `json:"state"`
	Authors        []string       `json:"authors"`
	Genres         []string       `json:"genres"`
	Languages      []string       `json:"languages"`
	ChaptersByLang map[string]int `json:"chapters_by_lang"`
	UpdatedAt      *time.Time     `json:"updated_at"`
}

// MangaListResponse is the paginated manga list response.
type MangaListResponse struct {
	Pagination
	Items []MangaSummary `json:"items"`
}

// MangaLangResponse holds per-language info in a manga detail response.
type MangaLangResponse struct {
	Lang             string     `json:"lang"`
	LatestUpdate     *time.Time `json:"latest_update"`
	TotalChapters    int        `json:"total_chapters"`
	FetchedChapters  int        `json:"fetched_chapters"`
	UploadedChapters int        `json:"uploaded_chapters"`
}

// MangaDetailResponse is the full detail response for a single manga.
type MangaDetailResponse struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	State       string              `json:"state"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Authors     []string            `json:"authors"`
	Genres      []string            `json:"genres"`
	CoverURL    string              `json:"cover_url"`
	Sources     map[string]string   `json:"sources"`
	Languages   []MangaLangResponse `json:"languages"`
	UpdatedAt   *time.Time          `json:"updated_at"`
}

// ChapterItem is one chapter entry in a chapter list response.
type ChapterItem struct {
	Chapter    string     `json:"chapter"`
	PageCount  int        `json:"page_count"`
	UploadedAt *time.Time `json:"uploaded_at"`
}

// ChapterListResponse is the list of uploaded chapters for a language.
type ChapterListResponse struct {
	ID       string        `json:"id"`
	Lang     string        `json:"lang"`
	Chapters []ChapterItem `json:"chapters"`
}

// ChapterReadResponse holds page URLs and prev/next navigation for a chapter.
type ChapterReadResponse struct {
	ID          string   `json:"id"`
	Lang        string   `json:"lang"`
	Chapter     string   `json:"chapter"`
	Pages       []string `json:"pages"`
	PrevChapter *string  `json:"prev_chapter"`
	NextChapter *string  `json:"next_chapter"`
}
