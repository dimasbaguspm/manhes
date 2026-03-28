package domain

// IngestRequested is published when a slug should be (re-)ingested.
// Published by: WatchlistService on Add and on each periodic tick.
type IngestRequested struct {
	DictionaryID string            `json:"dictionary_id"`
	Slug         string            `json:"slug"`
	Sources      map[string]string `json:"sources"`        // source_name → source_id
	LangToSource map[string]string `json:"lang_to_source"` // lang → source_name; empty means no assignment
}

// ChapterDownloaded is published after a chapter has been saved to disk.
// Published by: IngestService after each successful chapter download.
type ChapterDownloaded struct {
	DictionaryID string  `json:"dictionary_id"`
	Slug         string  `json:"slug"`
	Language     string  `json:"language"`
	ChapterNum   string  `json:"chapter_num"`
	SortKey      float64 `json:"sort_key"`
	PageCount    int     `json:"page_count"`
}

// ChapterUploaded is published after a chapter's pages have been uploaded to S3.
type ChapterUploaded struct {
	DictionaryID string   `json:"dictionary_id"`
	Slug         string   `json:"slug"`
	Language     string   `json:"language"`
	ChapterNum   string   `json:"chapter_num"`
	SortKey      float64  `json:"sort_key"`
	PageCount    int      `json:"page_count"`
	S3URLs       []string `json:"s3_urls"`
}

