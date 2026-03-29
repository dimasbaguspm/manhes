package application

import (
	"context"
	"log/slog"
	"time"

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
	s.log.Info("[Manga Service]: HandleDictionaryUpdated: received event",
		"dictionaryID", e.DictionaryID,
		"triggerIngest", e.TriggerIngest,
	)

	entry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Service]: GetDictionary failed", "dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[Manga Service]: dictionary entry not found, skipping", "dictionaryID", e.DictionaryID)
		return nil
	}
	s.log.Info("[Manga Service]: dictionary entry fetched",
		"dictionaryID", entry.ID,
		"slug", entry.Slug,
		"title", entry.Title,
		"sources", entry.Sources,
	)

	// Look up existing manga by dictionaryID to get its ID, or generate a new one.
	mangaID := ""
	existingManga, found, err := s.repo.GetMangaByDictionaryID(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Service]: GetMangaByDictionaryID failed", "dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if found {
		mangaID = existingManga.ID
		s.log.Info("[Manga Service]: existing manga found, reusing", "mangaID", mangaID, "dictionaryID", e.DictionaryID)
	} else {
		mangaID = uuid.NewString()
		s.log.Info("[Manga Service]: no existing manga, generating new ID", "mangaID", mangaID, "dictionaryID", e.DictionaryID)
	}

	// Derive chapters_by_lang from SourceStats: take the max count per language across sources.
	newChaptersByLang := s.computeChaptersByLang(entry.SourceStats)
	s.log.Info("[Manga Service]: chapters_by_lang derived from dictionary",
		"dictionaryID", entry.ID,
		"chaptersByLang", newChaptersByLang,
	)

	// Sync manga metadata from the best available source.
	ordered := s.registry.Ordered()
	s.log.Info("[Manga Service]: attempting sources", "sourceCount", len(ordered), "orderedSources", func() []string {
		names := make([]string, 0, len(ordered))
		for _, src := range ordered {
			names = append(names, src.Source())
		}
		return names
	}())

	mangaUpserted := false
	for _, src := range ordered {
		sourceID, ok := entry.Sources[src.Source()]
		if !ok {
			s.log.Debug("[Manga Service]: source not in dictionary entry, skipping", "source", src.Source())
			continue
		}
		s.log.Info("[Manga Service]: fetching manga detail", "source", src.Source(), "sourceID", sourceID)
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil {
			s.log.Warn("[Manga Service]: FetchMangaDetail error", "source", src.Source(), "sourceID", sourceID, "err", err)
			continue
		}
		if manga == nil {
			s.log.Warn("[Manga Service]: FetchMangaDetail returned nil manga", "source", src.Source(), "sourceID", sourceID)
			continue
		}

		// Always set state to fetching (chapters_by_lang is no longer stored on manga,
		// but newChaptersByLang is still used to drive the ChaptersFound event).
		manga.ID = mangaID
		manga.DictionaryID = entry.ID
		manga.CoverURL = entry.CoverURL
		manga.State = domain.StateFetching

		if err := s.repo.UpsertManga(ctx, *manga); err != nil {
			s.log.Error("[Manga Service]: UpsertManga failed", "mangaID", mangaID, "dictionaryID", entry.ID, "err", err)
			continue
		}
		s.log.Info("[Manga Service]: UpsertManga succeeded",
			"mangaID", manga.ID,
			"dictionaryID", entry.ID,
			"title", manga.Title,
			"state", manga.State,
		)
		mangaUpserted = true
		break
	}

	// If all FetchMangaDetail calls failed but we have a mangaID, upsert at least the manga metadata.
	if !mangaUpserted && mangaID != "" {
		s.log.Warn("[Manga Service]: no FetchMangaDetail succeeded, upserting with dictionary metadata only")
		m := existingManga.Manga
		m.ID = mangaID
		m.DictionaryID = entry.ID
		m.CoverURL = entry.CoverURL
		m.State = domain.StateFetching
		if err := s.repo.UpsertManga(ctx, m); err != nil {
			s.log.Error("[Manga Service]: fallback UpsertManga failed", "mangaID", mangaID, "err", err)
		} else {
			s.log.Info("[Manga Service]: fallback UpsertManga succeeded", "mangaID", mangaID)
			mangaUpserted = true
		}
	}

	// If TriggerIngest, publish ChaptersFound so the ingest pipeline fetches and downloads
	// chapters from sources. HandleChaptersFound will fetch page URLs for placeholder chapters
	// (those without PageURLs) directly from sources using the manga's dictionary entry.
	if e.TriggerIngest {
		s.log.Info("[Manga Service]: publishing ChaptersFound",
			"dictionaryID", entry.ID,
			"mangaID", mangaID,
			"chaptersByLang", newChaptersByLang,
		)
		if err := s.bus.Publish(ctx, s.cfg.Bus.ChaptersFound, domain.ChaptersFound{
			DictionaryID: entry.ID,
			MangaID:      mangaID,
			Chapters:     s.chaptersByLangToChapters(newChaptersByLang),
		}); err != nil {
			s.log.Error("[Manga Service]: publish ChaptersFound failed", "mangaID", mangaID, "dictionaryID", entry.ID, "err", err)
		}
	}

	return nil
}

// computeChaptersByLang derives chapter counts per language from SourceStats.
// For each language, it takes the maximum count across all sources (same logic as dictToResponse).
func (s *MangaService) computeChaptersByLang(stats map[string]domain.SourceStat) map[string]domain.ChapterStats {
	result := make(map[string]domain.ChapterStats)
	for _, count := range stats {
		if count.Err != "" {
			continue
		}
		for l, c := range count.ChaptersByLang {
			if c > result[l].Total {
				result[l] = domain.ChapterStats{Total: c}
			}
		}
	}
	return result
}

// HandleChapterUploaded is called after each chapter upload. Since manga no longer
// stores chapters_by_lang, this just logs the event. The final manga state is set
// by HandleMangaAvailable after all chapters are uploaded.
func (s *MangaService) HandleChapterUploaded(ctx context.Context, e domain.ChapterUploaded) error {
	s.log.Info("[Manga Service]: HandleChapterUploaded: received event",
		"mangaID", e.MangaID,
		"dictionaryID", e.DictionaryID,
		"language", e.Language,
		"chapterNum", e.ChapterNum,
	)
	// chapters_by_lang is no longer stored on manga; final state is computed by HandleMangaAvailable.
	return nil
}

// HandleMangaAvailable transitions a manga to the available state after all chapters have been uploaded.
func (s *MangaService) HandleMangaAvailable(ctx context.Context, e domain.MangaAvailable) error {
	s.log.Info("[Manga Service]: HandleMangaAvailable: received event",
		"mangaID", e.MangaID,
		"dictionaryID", e.DictionaryID,
	)

	mangaDetail, found, err := s.repo.GetMangaByDictionaryID(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Service]: HandleMangaAvailable: GetMangaByDictionaryID failed",
			"dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[Manga Service]: HandleMangaAvailable: manga not found, skipping",
			"dictionaryID", e.DictionaryID)
		return nil
	}

	// Get the dictionary to find expected chapter counts per language.
	dictEntry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Service]: HandleMangaAvailable: GetDictionary failed",
			"dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[Manga Service]: HandleMangaAvailable: dictionary not found, skipping",
			"dictionaryID", e.DictionaryID)
		return nil
	}

	// Compute expected chapters per language from SourceStats.
	expectedByLang := make(map[string]int)
	for _, stat := range dictEntry.SourceStats {
		if stat.Err != "" {
			continue
		}
		for lang, count := range stat.ChaptersByLang {
			if count > expectedByLang[lang] {
				expectedByLang[lang] = count
			}
		}
	}

	// Compute actual uploaded counts per language from the chapters table.
	uploadedByLang := make(map[string]int)
	storedChapters, err := s.repo.GetChaptersByManga(ctx, mangaDetail.ID)
	if err != nil {
		s.log.Error("[Manga Service]: HandleMangaAvailable: GetChaptersByManga failed",
			"mangaID", mangaDetail.ID, "err", err)
		return err
	}
	for _, ch := range storedChapters {
		uploaded, err := s.repo.GetChapterUploaded(ctx, mangaDetail.ID, ch.Language, ch.Number)
		if err != nil {
			s.log.Warn("[Manga Service]: HandleMangaAvailable: GetChapterUploaded failed",
				"mangaID", mangaDetail.ID, "lang", ch.Language, "chapter", ch.Number, "err", err)
			continue
		}
		if uploaded {
			uploadedByLang[ch.Language]++
		}
	}

	m := mangaDetail.Manga
	m.State = domain.StateAvailable

	if err := s.repo.UpsertManga(ctx, m); err != nil {
		s.log.Error("[Manga Service]: HandleMangaAvailable: UpsertManga failed",
			"mangaID", mangaDetail.ID, "err", err)
		return err
	}

	s.log.Info("[Manga Service]: HandleMangaAvailable: manga is now available",
		"mangaID", mangaDetail.ID,
		"dictionaryID", e.DictionaryID,
		"expectedByLang", expectedByLang,
		"uploadedByLang", uploadedByLang,
	)
	return nil
}

// chaptersByLangToChapters converts a chapters_by_lang map (lang→ChapterStats) into
// a map of language→[]Chapter suitable for publishing ChaptersFound.
// Each language gets one placeholder Chapter; HandleChaptersFound / HandleIngestRequested
// will replace these with real chapter objects fetched from sources.
func (s *MangaService) chaptersByLangToChapters(cbl map[string]domain.ChapterStats) map[string][]domain.Chapter {
	result := make(map[string][]domain.Chapter)
	for lang := range cbl {
		result[lang] = []domain.Chapter{{
			MangaID:   "",
			Number:    "",
			SortKey:   0,
			Title:     "",
			Language:  lang,
			Source:    "",
			SourceID:  "",
			PageURLs:  nil,
			ScrapedAt: time.Time{},
		}}
	}
	return result
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
