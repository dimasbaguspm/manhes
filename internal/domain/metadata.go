package domain

import "time"

// LanguageStat tracks chapter availability and download progress for one language.
type LanguageStat struct {
	Available  int `json:"available"`
	Downloaded int `json:"downloaded"`
}

// Metadata is the root source-of-truth written to library/{slug}/metadata.json.
type Metadata struct {
	Slug        string                  `json:"slug"`
	Title       string                  `json:"title"`
	Description string                  `json:"description,omitempty"`
	Status      string                  `json:"status"`
	Authors     []string                `json:"authors"`
	Genres      []string                `json:"genres"`
	Cover       string                  `json:"cover,omitempty"`
	Languages   map[string]LanguageStat `json:"languages"`
	Sources     map[string]string       `json:"sources"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// LangMetadata is written to library/{slug}/{lang}/metadata.json.
type LangMetadata struct {
	Slug       string    `json:"slug"`
	Language   string    `json:"language"`
	Available  int       `json:"available"`
	Downloaded int       `json:"downloaded"`
	Chapters   []float64 `json:"chapters"`
	UpdatedAt  time.Time `json:"updated_at"`
}
