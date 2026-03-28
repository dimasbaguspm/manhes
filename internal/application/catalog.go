package application

import (
	"context"

	"manga-engine/internal/domain"
)

var _ domain.CatalogQuerier = (*CatalogService)(nil)

type CatalogService struct {
	repo domain.Repository
}

func NewCatalogService(repo domain.Repository) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) ListManga(ctx context.Context, filter domain.MangaFilter) (domain.MangaPage, error) {
	return s.repo.ListManga(ctx, filter)
}

func (s *CatalogService) GetManga(ctx context.Context, dictionaryID string) (domain.MangaDetail, bool, error) {
	dictEntry, found, err := s.repo.GetDictionary(ctx, dictionaryID)
	if err != nil {
		return domain.MangaDetail{}, false, err
	}
	if !found {
		return domain.MangaDetail{}, false, nil
	}

	detail, found, err := s.repo.GetMangaBySlug(ctx, dictEntry.Slug)
	if err != nil {
		return domain.MangaDetail{}, false, err
	}
	if !found {
		// Manga not yet ingested — return partial detail from dictionary.
		return domain.MangaDetail{
			Manga: domain.Manga{
				DictionaryID: dictEntry.ID,
				Slug:         dictEntry.Slug,
				Title:        dictEntry.Title,
				Sources:      dictEntry.Sources,
				State:        dictEntry.State,
			},
		}, true, nil
	}

	detail.DictionaryID = dictEntry.ID
	detail.State = dictEntry.State
	detail.UpdatedAt = dictEntry.RefreshedAt
	applyLatestUpdates(&detail)
	return detail, true, nil
}

func (s *CatalogService) GetChaptersByLang(ctx context.Context, dictionaryID, lang string) ([]domain.MangaChapter, bool, error) {
	dictEntry, found, err := s.repo.GetDictionary(ctx, dictionaryID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	chapters, err := s.repo.GetChaptersByLang(ctx, dictEntry.Slug, lang)
	if err != nil {
		return nil, false, err
	}
	// Convert domain.Chapter to domain.MangaChapter
	result := make([]domain.MangaChapter, 0, len(chapters))
	for _, ch := range chapters {
		result = append(result, domain.MangaChapter{
			Slug:       ch.MangaSlug,
			Language:   ch.Language,
			ChapterNum: ch.Number,
			PageCount:  0,
			Uploaded:   false,
		})
	}
	return result, true, nil
}

func (s *CatalogService) ReadChapter(ctx context.Context, dictionaryID, lang string, num string) (domain.ChapterRead, bool, error) {
	dictEntry, found, err := s.repo.GetDictionary(ctx, dictionaryID)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if !found {
		return domain.ChapterRead{}, false, nil
	}

	chapters, err := s.repo.GetChaptersByLang(ctx, dictEntry.Slug, lang)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if len(chapters) == 0 {
		return domain.ChapterRead{}, false, nil
	}

	result := domain.ChapterRead{}
	for i, ch := range chapters {
		if ch.Number == num {
			// Build page URLs from image_src base path
			// For now return empty - actual page URLs would come from manifest
			if i > 0 {
				prev := chapters[i-1].Number
				result.PrevChapter = &prev
			}
			if i < len(chapters)-1 {
				next := chapters[i+1].Number
				result.NextChapter = &next
			}
			break
		}
	}
	return result, true, nil
}

func applyLatestUpdates(detail *domain.MangaDetail) {
	// Language stats are now derived from chapters table in the repository layer
}
