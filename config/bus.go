package config

// BusConfig holds event bus topic/routing constants.
type BusConfig struct {
	IngestRequested   string // "ingest.requested"
	ChaptersFound     string // "chapters.found"
	ChapterUploaded   string // "chapter.uploaded"
	ChapterDownloaded string // "chapter.downloaded"
	DictionaryUpdated string // "dictionary.updated"
}

func loadBusConfig() BusConfig {
	return BusConfig{
		IngestRequested:   "ingest.requested",
		ChaptersFound:     "chapters.found",
		ChapterUploaded:   "chapter.uploaded",
		ChapterDownloaded: "chapter.downloaded",
		DictionaryUpdated: "dictionary.updated",
	}
}
