package domain

import (
	"context"
	"time"
)

// ChapterStats holds per-language chapter counts.
type ChapterStats struct {
	Total     int `json:"total"`
	Available int `json:"available"`
}

// Manga is the core aggregate representing a manga title.
// It has a 1:1 relationship with DictionaryEntry via DictionaryID.
type Manga struct {
	ID           string       `json:"id"`
	DictionaryID string       `json:"dictionary_id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       string       `json:"status"` // ongoing | completed | hiatus
	Authors      []string     `json:"authors"`
	Genres       []string     `json:"genres"`
	CoverURL     string       `json:"cover_url"`
	State        MangaState   `json:"state"`
	UpdatedAt    *time.Time   `json:"updated_at"`
	CreatedAt    time.Time    `json:"created_at"`
}

// MangaFilter holds optional filters and pagination for ListManga.
type MangaFilter struct {
	IDs       []string // filter by dictionary_id array
	Q         string   // title OR description search
	Genres    []string // genre array (OR between items)
	Authors   []string // author array (OR between items)
	States    []string // state array: unavailable, fetching, available
	SortBy    string   // title | updatedAt | createdAt (default: title)
	SortOrder string   // asc | desc (default: asc)
	Page      int      // 1-based (default: 1)
	PageSize  int      // (default: 20, max: 100)
}

// MangaPage is a paginated slice of Manga.
type MangaPage struct {
	Items    []Manga `json:"items"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

// ChapterRef is a lightweight reference to a chapter pending S3 upload.
type ChapterRef struct {
	DictionaryID string // used as S3 path prefix
	MangaID      string
	Language     string
	ChapterNum   string
}

// MangaDetail includes full manga info with languages and chapters.
type MangaDetail struct {
	Manga
	Languages []MangaLang    `json:"languages"`
	Chapters  []MangaChapter `json:"chapters"`
}

// MangaLang holds per-language availability stats.
type MangaLang struct {
	Language     string     `json:"language"`
	Total        int        `json:"total"`
	Available    int        `json:"available"`
	Fetched      int        `json:"fetched"`
	LatestUpdate *time.Time `json:"latest_update"`
}

// MangaChapter holds chapter catalog metadata.
type MangaChapter struct {
	MangaID    string     `json:"manga_id"`
	Language   string     `json:"language"`
	ID         string     `json:"id"`
	Order      int        `json:"order"`
	Name       string     `json:"name"`
	UpdatedAt  *time.Time `json:"updated_at"`
	PageCount  int        `json:"page_count"`
	Uploaded   bool       `json:"uploaded"`
}

// ChapterRead holds the result of reading a chapter: page URLs and prev/next navigation.
type ChapterRead struct {
	MangaID     string   `json:"manga_id"`
	Pages       []string `json:"pages"`
	PrevChapter *string  `json:"prev_chapter"`
	NextChapter *string  `json:"next_chapter"`
}

// MangaQuerier is the read-only manga catalog port used by the HTTP handler.
type MangaQuerier interface {
	ListManga(ctx context.Context, filter MangaFilter) (MangaPage, error)
	GetManga(ctx context.Context, mangaID string) (MangaDetail, bool, error)
	GetMangaLanguages(ctx context.Context, mangaID, dictionaryID string) ([]MangaLangResponse, error)
	GetChaptersByLang(ctx context.Context, mangaID, lang string) ([]MangaChapter, bool, error)
	ReadChapter(ctx context.Context, chapterID string) (ChapterRead, bool, error)
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
	ID           string              `json:"id"`
	DictionaryID string              `json:"dictionary_id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Status       string              `json:"status"`
	CoverURL     string              `json:"cover_url"`
	State        string              `json:"state"`
	Authors      []string            `json:"authors"`
	Genres       []string            `json:"genres"`
	Languages    []MangaLangResponse `json:"languages"`
	UpdatedAt    *time.Time          `json:"updated_at"`
	CreatedAt    time.Time           `json:"created_at"`
}

// MangaListResponse is the paginated manga list response.
type MangaListResponse struct {
	Pagination
	Items []MangaSummary `json:"items"`
}

// MangaLangResponse holds per-language info in a manga detail response.
type MangaLangResponse struct {
	Lang              string     `json:"lang"`
	TotalChapters     int        `json:"total_chapters"`
	AvailableChapters int        `json:"available_chapters"`
	LatestUpdate      *time.Time `json:"latest_update"`
}

// MangaDetailResponse is the full detail response for a single manga.
type MangaDetailResponse struct {
	ID           string              `json:"id"`
	DictionaryID string              `json:"dictionary_id"`
	Title        string              `json:"title"`
	State        string              `json:"state"`
	Description  string              `json:"description"`
	Status       string              `json:"status"`
	Authors      []string            `json:"authors"`
	Genres       []string            `json:"genres"`
	CoverURL     string              `json:"cover_url"`
	Languages    []MangaLangResponse `json:"languages"`
	UpdatedAt    *time.Time          `json:"updated_at"`
	CreatedAt    time.Time           `json:"created_at"`
}

// ChapterItem is one chapter entry in a chapter list response.
type ChapterItem struct {
	ID        string     `json:"id"`
	Order     int        `json:"order"`
	Name      string     `json:"name"`
	UpdatedAt *time.Time `json:"updated_at"`
	PageCount int        `json:"page_count"`
}

// ChapterListResponse is the list of uploaded chapters for a language.
type ChapterListResponse struct {
	ID       string        `json:"id"`
	Lang     string        `json:"lang"`
	Chapters []ChapterItem `json:"chapters"`
}

// ChapterReadResponse holds page URLs and prev/next navigation for a chapter.
type ChapterReadResponse struct {
	MangaID     string   `json:"manga_id"`
	ChapterID   string   `json:"chapter_id"`
	Pages       []string `json:"pages"`
	PrevChapter *string  `json:"prev_chapter"`
	NextChapter *string  `json:"next_chapter"`
}
