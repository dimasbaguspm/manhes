package domain

import "context"

// MangaSubscriber handles manga-related domain events.
type MangaSubscriber interface {
	HandleDictionaryUpdated(ctx context.Context, e DictionaryUpdated) error
	HandleChapterUploaded(ctx context.Context, e ChapterUploaded) error
	HandleMangaAvailable(ctx context.Context, e MangaAvailable) error
}

// DictionarySubscriber handles dictionary-related domain events.
type DictionarySubscriber interface {
	HandleDictionaryRefreshed(ctx context.Context, e DictionaryRefreshed) error
}
