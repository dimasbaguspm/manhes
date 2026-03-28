package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.MangaManager = (*MangaService)(nil)
var _ domain.MangaQuerier = (*MangaService)(nil)

type MangaServiceConfig struct {
	Repo     domain.Repository
	Registry domain.SourceRegistry
	Bus      domain.EventBus
	Cfg      *config.Config
}

type MangaService struct {
	repo     domain.Repository
	registry domain.SourceRegistry
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewMangaService(cfg MangaServiceConfig) *MangaService {
	return &MangaService{
		repo:     cfg.Repo,
		registry: cfg.Registry,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("service", "manga"),
	}
}

func (s *MangaService) HandleDictionaryUpdated(ctx context.Context, e domain.DictionaryUpdated) error {
	entry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	// Look up existing manga by dictionaryID to get its ID, or generate a new one.
	mangaID := ""
	existing, found, err := s.repo.GetMangaByDictionaryID(ctx, e.DictionaryID)
	if err != nil {
		return err
	}
	if found {
		mangaID = existing.ID
	} else {
		mangaID = uuid.NewString()
	}

	// Sync manga metadata from the best available source.
	ordered := s.registry.Ordered()
	for _, src := range ordered {
		sourceID, ok := entry.Sources[src.Source()]
		if !ok {
			continue
		}
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil || manga == nil {
			continue
		}
		manga.ID = mangaID
		manga.DictionaryID = entry.ID
		manga.CoverURL = entry.CoverURL
		if err := s.repo.UpsertManga(ctx, *manga); err != nil {
			s.log.Warn("manga service: upsert manga", "mangaID", mangaID, "dictionaryID", entry.ID, "err", err)
		}
		break
	}

	// If TriggerIngest, also publish ingest.requested so chapters are synced.
	if e.TriggerIngest {
		sources := s.sourcesForIngest(entry)
		if err := s.bus.Publish(ctx, s.cfg.Bus.IngestRequested, domain.IngestRequested{
			DictionaryID: entry.ID,
			MangaID:      mangaID,
			Sources:      sources,
			LangToSource: entry.BestSource,
		}); err != nil {
			s.log.Warn("manga service: publish ingest.requested", "mangaID", mangaID, "dictionaryID", entry.ID, "err", err)
		}
	}

	return nil
}

// sourcesForIngest returns only sources that are best for at least one language,
// avoiding redundant downloads when BestSource is populated.
// Falls back to all sources before the first dictionary refresh has run.
func (s *MangaService) sourcesForIngest(entry domain.DictionaryEntry) map[string]string {
	if len(entry.Sources) == 0 {
		return map[string]string{}
	}
	if len(entry.BestSource) > 0 {
		result := make(map[string]string)
		for _, sourceName := range entry.BestSource {
			if id, ok := entry.Sources[sourceName]; ok {
				result[sourceName] = id
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return entry.Sources
}

// ListManga returns a paginated list of manga matching the given filter.
func (s *MangaService) ListManga(ctx context.Context, filter domain.MangaFilter) (domain.MangaPage, error) {
	return s.repo.ListManga(ctx, filter)
}

// GetManga returns manga detail for a given dictionary ID.
// Returns false if the manga has not yet been ingested.
func (s *MangaService) GetManga(ctx context.Context, dictionaryID string) (domain.MangaDetail, bool, error) {
	detail, found, err := s.repo.GetMangaByDictionaryID(ctx, dictionaryID)
	if err != nil {
		return domain.MangaDetail{}, false, err
	}
	if found {
		return detail, true, nil
	}
	return domain.MangaDetail{}, false, nil
}

// GetChaptersByLang returns uploaded chapters for a given dictionary ID and language.
func (s *MangaService) GetChaptersByLang(ctx context.Context, dictionaryID, lang string) ([]domain.MangaChapter, bool, error) {
	manga, found, err := s.repo.GetMangaByDictionaryID(ctx, dictionaryID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	chapters, err := s.repo.GetChaptersByLang(ctx, manga.ID, lang)
	if err != nil {
		return nil, false, err
	}
	result := make([]domain.MangaChapter, 0, len(chapters))
	for _, ch := range chapters {
		result = append(result, domain.MangaChapter{
			MangaID:    ch.MangaID,
			Language:   ch.Language,
			ChapterNum: ch.Number,
			PageCount:  0,
			Uploaded:   false,
		})
	}
	return result, true, nil
}

// ReadChapter returns chapter read info (pages + prev/next navigation) for a given dictionary ID, language, and chapter number.
func (s *MangaService) ReadChapter(ctx context.Context, dictionaryID, lang string, num string) (domain.ChapterRead, bool, error) {
	manga, found, err := s.repo.GetMangaByDictionaryID(ctx, dictionaryID)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if !found {
		return domain.ChapterRead{}, false, nil
	}

	chapters, err := s.repo.GetChaptersByLang(ctx, manga.ID, lang)
	if err != nil {
		return domain.ChapterRead{}, false, err
	}
	if len(chapters) == 0 {
		return domain.ChapterRead{}, false, nil
	}

	result := domain.ChapterRead{}
	for i, ch := range chapters {
		if ch.Number == num {
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
