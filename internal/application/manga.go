package application

import (
	"context"
	"log/slog"
	"time"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.MangaQuerier = (*MangaService)(nil)

type MangaServiceConfig struct {
	Repo domain.Repository
	Cfg  *config.Config
}

type MangaService struct {
	repo domain.Repository
	cfg  *config.Config
	log  *slog.Logger
}

func NewMangaService(cfg MangaServiceConfig) *MangaService {
	return &MangaService{
		repo: cfg.Repo,
		cfg:  cfg.Cfg,
		log:  slog.With("service", "manga"),
	}
}

// ListManga returns a paginated list of manga matching the given filter.
func (s *MangaService) ListManga(ctx context.Context, filter domain.MangaFilter) (domain.MangaPage, error) {
	return s.repo.ListManga(ctx, filter)
}

// GetMangaLanguages returns per-language stats for a manga.
func (s *MangaService) GetMangaLanguages(ctx context.Context, mangaID, dictionaryID string) ([]domain.MangaLangResponse, error) {
	return s.repo.GetMangaLanguages(ctx, mangaID, dictionaryID)
}

// GetManga returns manga detail for a given manga ID.
// Returns false if the manga has not yet been ingested.
func (s *MangaService) GetManga(ctx context.Context, mangaID string) (domain.MangaDetail, bool, error) {
	detail, found, err := s.repo.GetMangaByID(ctx, mangaID)
	if err != nil {
		return domain.MangaDetail{}, false, err
	}
	if found {
		return detail, true, nil
	}
	return domain.MangaDetail{}, false, nil
}

// GetChaptersByLang returns uploaded chapters for a given manga ID and language.
func (s *MangaService) GetChaptersByLang(ctx context.Context, mangaID, lang string) ([]domain.MangaChapter, bool, error) {
	_, found, err := s.repo.GetMangaByID(ctx, mangaID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	chapters, err := s.repo.GetUploadedChaptersByLang(ctx, mangaID, lang)
	if err != nil {
		return nil, false, err
	}
	result := make([]domain.MangaChapter, 0, len(chapters))
	for _, ch := range chapters {
		var updatedAt *time.Time
		if !ch.UpdatedAt.IsZero() {
			ua := ch.UpdatedAt
			updatedAt = &ua
		}
		result = append(result, domain.MangaChapter{
			MangaID:   ch.MangaID,
			Language:  ch.Language,
			ID:        ch.ID,
			Order:     int(ch.SortKey),
			Name:      ch.Number,
			UpdatedAt: updatedAt,
			PageCount: ch.PageCount,
			Uploaded:  len(ch.PageURLs) > 0,
		})
	}
	return result, true, nil
}

// ReadChapter returns chapter read info (pages + prev/next navigation) for a given chapter ID.
func (s *MangaService) ReadChapter(ctx context.Context, chapterID string) (domain.ChapterRead, bool, error) {
	// Get chapter directly by ID.
	chapter, err := s.repo.GetChapterByID(ctx, chapterID)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if chapter == nil {
		return domain.ChapterRead{}, false, nil
	}

	// Get all chapters for this manga/lang to compute prev/next.
	allChapters, err := s.repo.GetChaptersByLang(ctx, chapter.MangaID, chapter.Language)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}

	result := domain.ChapterRead{
		MangaID: chapter.MangaID,
		Pages:   chapter.PageURLs,
	}
	for i, ch := range allChapters {
		if ch.ID == chapterID {
			if i > 0 {
				prev := allChapters[i-1].ID
				result.PrevChapter = &prev
			}
			if i < len(allChapters)-1 {
				next := allChapters[i+1].ID
				result.NextChapter = &next
			}
			break
		}
	}
	return result, true, nil
}
