package handler

import (
	"context"

	"manga-engine/internal/domain"
)

func (h *Handlers) ListManga(ctx context.Context, filter domain.MangaFilter) (domain.MangaPage, []domain.MangaSummary, error) {
	result, err := h.Repo.ListManga(ctx, filter)
	if err != nil {
		return domain.MangaPage{}, nil, err
	}

	items := make([]domain.MangaSummary, 0, len(result.Items))
	for _, m := range result.Items {
		languages, err := h.Repo.GetMangaLanguages(ctx, m.ID, m.DictionaryID)
		if err != nil {
			return domain.MangaPage{}, nil, err
		}
		items = append(items, toMangaSummary(m, languages))
	}

	return result, items, nil
}

func (h *Handlers) GetManga(ctx context.Context, mangaID string) (domain.MangaDetail, bool, error) {
	return h.Repo.GetMangaByID(ctx, mangaID)
}

func (h *Handlers) GetChaptersByLang(ctx context.Context, mangaID, lang string) ([]domain.ChapterItem, bool, error) {
	_, found, err := h.Repo.GetMangaByID(ctx, mangaID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	chapters, err := h.Repo.GetUploadedChaptersByLang(ctx, mangaID, lang)
	if err != nil {
		return nil, false, err
	}

	items := make([]domain.ChapterItem, 0, len(chapters))
	for _, ch := range chapters {
		items = append(items, toChapterItem(ch))
	}

	return items, true, nil
}
