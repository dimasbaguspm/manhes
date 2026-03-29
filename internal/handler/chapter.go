package handler

import (
	"context"

	"manga-engine/internal/domain"
)

func (h *Handlers) ReadChapter(ctx context.Context, chapterID string) (domain.ChapterRead, bool, error) {
	chapter, err := h.Repo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if chapter == nil {
		return domain.ChapterRead{}, false, nil
	}

	allChapters, err := h.Repo.GetChaptersByLang(ctx, chapter.MangaID, chapter.Language)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}

	result := domain.ChapterRead{
		MangaID: chapter.MangaID,
		Pages:   chapter.PageURLs,
	}
	prev, next := buildPrevNext(allChapters, chapterID)
	result.PrevChapter = prev
	result.NextChapter = next

	return result, true, nil
}
