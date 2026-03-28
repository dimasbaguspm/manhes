package domain

// IngestRequested is published when a slug should be (re-)ingested.
// Published by: MangaService when DictionaryUpdated has TriggerIngest=true.
type IngestRequested struct {
	DictionaryID string            `json:"dictionary_id"`
	MangaID      string            `json:"manga_id"`       // manga UUID; empty means create new
	Sources      map[string]string `json:"sources"`        // source_name → source_id
	LangToSource map[string]string `json:"lang_to_source"` // lang → source_name; empty means no assignment
}

// ChaptersFound is published after fetchFromSources completes for a manga.
// It carries the full chapter list per language discovered from sources, filtered
// to only include chapters not already in the database.
type ChaptersFound struct {
	DictionaryID string               `json:"dictionary_id"`
	MangaID      string               `json:"manga_id"`
	Chapters     map[string][]Chapter `json:"chapters"` // lang -> chapters
}

// ChapterDownloaded is published after a chapter has been saved to disk.
// Published by: IngestService after each successful chapter download.
type ChapterDownloaded struct {
	DictionaryID string  `json:"dictionary_id"`
	MangaID      string  `json:"manga_id"`
	Language     string  `json:"language"`
	ChapterNum   string  `json:"chapter_num"`
	SortKey      float64 `json:"sort_key"`
	PageCount    int     `json:"page_count"`
}

// ChapterUploaded is published after a chapter's pages have been uploaded to S3.
type ChapterUploaded struct {
	DictionaryID string   `json:"dictionary_id"`
	MangaID      string   `json:"manga_id"`
	Language     string   `json:"language"`
	ChapterNum   string   `json:"chapter_num"`
	SortKey      float64  `json:"sort_key"`
	PageCount    int      `json:"page_count"`
	S3URLs       []string `json:"s3_urls"`
}

// DictionaryUpdated is published after a dictionary entry is refreshed and
// its cross-source data (Sources, BestSource, SourceStats) may have changed.
// When TriggerIngest is true, the subscriber should also publish ingest.requested
// to sync chapters (set by Refresh, not by Search).
type DictionaryUpdated struct {
	DictionaryID  string `json:"dictionary_id"`
	TriggerIngest bool   `json:"trigger_inggest"` // true when fired from Refresh
}
