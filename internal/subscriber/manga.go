package subscriber

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

// MangaSubscriberConfig holds dependencies for MangaSubscriber.
type MangaSubscriberConfig struct {
	Repo     domain.Repository
	Registry domain.SourceRegistry
	Bus      domain.EventBus
	Cfg      *config.Config
}

// MangaSubscriber handles manga-related events.
type MangaSubscriber struct {
	repo     domain.Repository
	registry domain.SourceRegistry
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewMangaSubscriber(cfg MangaSubscriberConfig) *MangaSubscriber {
	return &MangaSubscriber{
		repo:     cfg.Repo,
		registry: cfg.Registry,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("component", "subscriber"),
	}
}

func (s *MangaSubscriber) HandleDictionaryUpdated(ctx context.Context, e domain.DictionaryUpdated) error {
	s.log.Info("[Manga Subscriber]: HandleDictionaryUpdated: received event",
		"dictionaryID", e.DictionaryID,
		"triggerIngest", e.TriggerIngest,
	)

	entry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Subscriber]: GetDictionary failed", "dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[Manga Subscriber]: dictionary entry not found, skipping", "dictionaryID", e.DictionaryID)
		return nil
	}
	s.log.Info("[Manga Subscriber]: dictionary entry fetched",
		"dictionaryID", entry.ID,
		"slug", entry.Slug,
		"title", entry.Title,
		"sources", entry.Sources,
	)

	// Look up existing manga by dictionaryID to get its ID, or generate a new one.
	mangaID := ""
	existingManga, found, err := s.repo.GetMangaByDictionaryID(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Subscriber]: GetMangaByDictionaryID failed", "dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if found {
		mangaID = existingManga.ID
		s.log.Info("[Manga Subscriber]: existing manga found, reusing", "mangaID", mangaID, "dictionaryID", e.DictionaryID)
	} else {
		mangaID = uuid.NewString()
		s.log.Info("[Manga Subscriber]: no existing manga, generating new ID", "mangaID", mangaID, "dictionaryID", e.DictionaryID)
	}

	// Derive chapters_by_lang from SourceStats: take the max count per language across sources.
	newChaptersByLang := s.computeChaptersByLang(entry.SourceStats)
	s.log.Info("[Manga Subscriber]: chapters_by_lang derived from dictionary",
		"dictionaryID", entry.ID,
		"chaptersByLang", newChaptersByLang,
	)

	// Sync manga metadata from the best available source.
	ordered := s.registry.Ordered()
	s.log.Info("[Manga Subscriber]: attempting sources", "sourceCount", len(ordered), "orderedSources", func() []string {
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
			s.log.Debug("[Manga Subscriber]: source not in dictionary entry, skipping", "source", src.Source())
			continue
		}
		s.log.Info("[Manga Subscriber]: fetching manga detail", "source", src.Source(), "sourceID", sourceID)
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil {
			s.log.Warn("[Manga Subscriber]: FetchMangaDetail error", "source", src.Source(), "sourceID", sourceID, "err", err)
			continue
		}
		if manga == nil {
			s.log.Warn("[Manga Subscriber]: FetchMangaDetail returned nil manga", "source", src.Source(), "sourceID", sourceID)
			continue
		}

		manga.ID = mangaID
		manga.DictionaryID = entry.ID
		manga.CoverURL = entry.CoverURL
		manga.State = domain.StateFetching

		if err := s.repo.UpsertManga(ctx, *manga); err != nil {
			s.log.Error("[Manga Subscriber]: UpsertManga failed", "mangaID", mangaID, "dictionaryID", entry.ID, "err", err)
			continue
		}
		s.log.Info("[Manga Subscriber]: UpsertManga succeeded",
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
		s.log.Warn("[Manga Subscriber]: no FetchMangaDetail succeeded, upserting with dictionary metadata only")
		m := existingManga.Manga
		m.ID = mangaID
		m.DictionaryID = entry.ID
		m.CoverURL = entry.CoverURL
		m.State = domain.StateFetching
		if err := s.repo.UpsertManga(ctx, m); err != nil {
			s.log.Error("[Manga Subscriber]: fallback UpsertManga failed", "mangaID", mangaID, "err", err)
		} else {
			s.log.Info("[Manga Subscriber]: fallback UpsertManga succeeded", "mangaID", mangaID)
			mangaUpserted = true
		}
	}

	// If TriggerIngest, publish ChaptersFound so the ingest pipeline fetches and downloads
	// chapters from sources.
	if e.TriggerIngest {
		s.log.Info("[Manga Subscriber]: publishing ChaptersFound",
			"dictionaryID", entry.ID,
			"mangaID", mangaID,
			"chaptersByLang", newChaptersByLang,
		)
		if err := s.bus.Publish(ctx, s.cfg.Bus.ChaptersFound, domain.ChaptersFound{
			DictionaryID: entry.ID,
			MangaID:      mangaID,
			Chapters:     s.chaptersByLangToChapters(newChaptersByLang),
		}); err != nil {
			s.log.Error("[Manga Subscriber]: publish ChaptersFound failed", "mangaID", mangaID, "dictionaryID", entry.ID, "err", err)
		}
	}

	return nil
}

// computeChaptersByLang derives chapter counts per language from SourceStats.
func (s *MangaSubscriber) computeChaptersByLang(stats map[string]domain.SourceStat) map[string]domain.ChapterStats {
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

// chaptersByLangToChapters converts a chapters_by_lang map into a map of language→[]Chapter.
func (s *MangaSubscriber) chaptersByLangToChapters(cbl map[string]domain.ChapterStats) map[string][]domain.Chapter {
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

// HandleChapterUploaded is called after each chapter upload.
func (s *MangaSubscriber) HandleChapterUploaded(ctx context.Context, e domain.ChapterUploaded) error {
	s.log.Info("[Manga Subscriber]: HandleChapterUploaded: received event",
		"mangaID", e.MangaID,
		"dictionaryID", e.DictionaryID,
		"language", e.Language,
		"chapterNum", e.ChapterNum,
	)
	return nil
}

// HandleMangaAvailable transitions a manga to the available state after all chapters have been uploaded.
func (s *MangaSubscriber) HandleMangaAvailable(ctx context.Context, e domain.MangaAvailable) error {
	s.log.Info("[Manga Subscriber]: HandleMangaAvailable: received event",
		"mangaID", e.MangaID,
		"dictionaryID", e.DictionaryID,
	)

	mangaDetail, found, err := s.repo.GetMangaByDictionaryID(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Subscriber]: HandleMangaAvailable: GetMangaByDictionaryID failed",
			"dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[Manga Subscriber]: HandleMangaAvailable: manga not found, skipping",
			"dictionaryID", e.DictionaryID)
		return nil
	}

	// Get the dictionary to find expected chapter counts per language.
	dictEntry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Manga Subscriber]: HandleMangaAvailable: GetDictionary failed",
			"dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[Manga Subscriber]: HandleMangaAvailable: dictionary not found, skipping",
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
		s.log.Error("[Manga Subscriber]: HandleMangaAvailable: GetChaptersByManga failed",
			"mangaID", mangaDetail.ID, "err", err)
		return err
	}
	for _, ch := range storedChapters {
		uploaded, err := s.repo.GetChapterUploaded(ctx, mangaDetail.ID, ch.Language, ch.Number)
		if err != nil {
			s.log.Warn("[Manga Subscriber]: HandleMangaAvailable: GetChapterUploaded failed",
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
		s.log.Error("[Manga Subscriber]: HandleMangaAvailable: UpsertManga failed",
			"mangaID", mangaDetail.ID, "err", err)
		return err
	}

	s.log.Info("[Manga Subscriber]: HandleMangaAvailable: manga is now available",
		"mangaID", mangaDetail.ID,
		"dictionaryID", e.DictionaryID,
		"expectedByLang", expectedByLang,
		"uploadedByLang", uploadedByLang,
	)
	return nil
}
