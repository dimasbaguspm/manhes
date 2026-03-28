package application

import (
	"context"

	"golang.org/x/sync/errgroup"

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
	return chapters, true, err
}

func (s *CatalogService) ReadChapter(ctx context.Context, dictionaryID, lang string, num string) (domain.ChapterRead, bool, error) {
	dictEntry, found, err := s.repo.GetDictionary(ctx, dictionaryID)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if !found {
		return domain.ChapterRead{}, false, nil
	}

	var (
		pages    []string
		chapters []domain.MangaChapter
		g        errgroup.Group
	)
	g.Go(func() error {
		var err error
		pages, err = s.repo.GetChapterPages(ctx, dictEntry.Slug, lang, num)
		return err
	})
	g.Go(func() error {
		var err error
		chapters, err = s.repo.GetChaptersByLang(ctx, dictEntry.Slug, lang)
		return err
	})
	if err := g.Wait(); err != nil {
		return domain.ChapterRead{}, false, err
	}
	if len(pages) == 0 {
		return domain.ChapterRead{}, false, nil
	}

	result := domain.ChapterRead{Pages: pages}
	for i, ch := range chapters {
		if ch.ChapterNum == num {
			if i > 0 {
				prev := chapters[i-1].ChapterNum
				result.PrevChapter = &prev
			}
			if i < len(chapters)-1 {
				next := chapters[i+1].ChapterNum
				result.NextChapter = &next
			}
			break
		}
	}
	return result, true, nil
}

func applyLatestUpdates(detail *domain.MangaDetail) {
	byLang := map[string]*domain.MangaLang{}
	for i := range detail.Languages {
		byLang[detail.Languages[i].Language] = &detail.Languages[i]
	}
	for _, ch := range detail.Chapters {
		entry, ok := byLang[ch.Language]
		if !ok {
			continue
		}
		if ch.Uploaded {
			entry.Uploaded++
		}
		if ch.UploadedAt != nil {
			if entry.LatestUpdate == nil || ch.UploadedAt.After(*entry.LatestUpdate) {
				entry.LatestUpdate = ch.UploadedAt
			}
		}
	}
}
