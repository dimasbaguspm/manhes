package config

// BusConfig holds event bus topic/routing constants.
type BusConfig struct {
	IngestRequested  string // "ingest.requested"
	ChapterUploaded  string // "chapter.uploaded"
	ChapterDownloaded string // "chapter.downloaded"
}

func loadBusConfig() BusConfig {
	return BusConfig{
		IngestRequested:   "ingest.requested",
		ChapterUploaded:   "chapter.uploaded",
		ChapterDownloaded: "chapter.downloaded",
	}
}
