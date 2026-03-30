package handler

import (
	"context"
	"mime"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/storage"
	"manga-engine/pkg/concurrent"
)

type searchHit struct {
	slug        string
	title       string
	sources     map[string]string
	sourceStats map[string]domain.SourceStat
}

func (h *Handlers) Search(ctx context.Context, query string) ([]domain.DictionaryEntry, error) {
	h.Log.Info("[DictionaryService] Search: started", "query", query)

	hits := h.searchAllSources(ctx, query)

	existingIDs := make(map[string]string, len(hits))
	for slug := range hits {
		entry, found, err := h.Repo.GetDictionaryBySlug(ctx, slug)
		if err != nil {
			h.Log.Warn("search: GetDictionaryBySlug failed", "slug", slug, "err", err)
			continue
		}
		if found {
			existingIDs[slug] = entry.ID
		}
	}

	entries := h.buildEntries(hits, existingIDs)
	if err := h.Repo.UpsertDictionaryBatch(ctx, entries); err != nil {
		return nil, err
	}
	h.logBatchUpsert(entries)

	for _, e := range entries {
		if err := h.Bus.Publish(ctx, h.Cfg.Bus.DictionaryUpdated, domain.DictionaryUpdated{
			DictionaryID:  e.ID,
			TriggerIngest: false,
		}); err != nil {
			h.Log.Warn("search: publish dictionary.updated", "id", e.ID, "err", err)
		}
	}
	return entries, nil
}

func (h *Handlers) Refresh(ctx context.Context, id string) (domain.DictionaryEntry, error) {
	h.Log.Info("[DictionaryService] Refresh: started", "dictionaryID", id)

	entry, found, err := h.Repo.GetDictionary(ctx, id)
	if err != nil {
		h.Log.Error("[DictionaryService] Refresh: GetDictionary failed", "dictionaryID", id, "err", err)
		return domain.DictionaryEntry{}, err
	}
	if !found {
		h.Log.Warn("[DictionaryService] Refresh: entry not found", "dictionaryID", id)
		return domain.DictionaryEntry{}, domain.ErrNotFound
	}

	discovered := h.discoverSourcesForEntry(ctx, entry)
	oldSources := entry.Sources
	h.mergeDiscoveredSources(entry, discovered)

	if _, err := h.refreshWithOldState(ctx, id, oldSources); err != nil {
		return domain.DictionaryEntry{}, err
	}

	updated, found, err := h.Repo.GetDictionary(ctx, id)
	if err != nil || !found {
		return domain.DictionaryEntry{}, err
	}

	if err := h.Bus.Publish(ctx, h.Cfg.Bus.DictionaryUpdated, domain.DictionaryUpdated{
		DictionaryID:  updated.ID,
		TriggerIngest: true,
	}); err != nil {
		h.Log.Warn("[DictionaryService] Refresh: publish DictionaryUpdated failed", "dictionaryID", id, "err", err)
	}

	h.Log.Info("[DictionaryService] Refresh: completed", "dictionaryID", id)
	return updated, nil
}

// searchAllSources runs concurrent searches across all scrapers and merges results by slug.
func (h *Handlers) searchAllSources(ctx context.Context, query string) map[string]*searchHit {
	hits := make(map[string]*searchHit)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, src := range h.Registry.Ordered() {
		searcher, ok := src.(domain.Searcher)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(searcher domain.Searcher, scraperSrc domain.Scraper) {
			defer wg.Done()
			results, err := searcher.Search(ctx, query)
			if err != nil {
				h.Log.Warn("search: source failed", "source", scraperSrc.Source(), "err", err)
				return
			}
			h.Log.Info("search: source results", "source", scraperSrc.Source(), "query", query, "count", len(results))
			for _, r := range results {
				slug := storage.Slugify(r.Title)
				stat := h.fetchSourceStat(ctx, scraperSrc, r.ID)

				mu.Lock()
				if hit, ok := hits[slug]; ok {
					hit.sources[scraperSrc.Source()] = r.ID
					hit.sourceStats[scraperSrc.Source()] = stat
				} else {
					hits[slug] = &searchHit{
						slug:        slug,
						title:       r.Title,
						sources:     map[string]string{scraperSrc.Source(): r.ID},
						sourceStats: map[string]domain.SourceStat{scraperSrc.Source(): stat},
					}
				}
				mu.Unlock()
			}
		}(searcher, src)
	}
	wg.Wait()
	return hits
}

// discoverSourcesForEntry searches all scrapers for entries matching the given entry's title/slug.
func (h *Handlers) discoverSourcesForEntry(ctx context.Context, entry domain.DictionaryEntry) []struct {
	name string
	id   string
} {
	var discovered []struct {
		name string
		id   string
	}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, src := range h.Registry.Ordered() {
		searcher, ok := src.(domain.Searcher)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(searcher domain.Searcher, scraperSrc domain.Scraper) {
			defer wg.Done()
			results, err := searcher.Search(ctx, entry.Title)
			if err != nil {
				h.Log.Warn("dictionary refresh: search failed", "source", scraperSrc.Source(), "title", entry.Title, "err", err)
				return
			}
			for _, r := range results {
				if storage.Slugify(r.Title) == entry.Slug {
					mu.Lock()
					discovered = append(discovered, struct {
						name string
						id   string
					}{scraperSrc.Source(), r.ID})
					mu.Unlock()
				}
			}
		}(searcher, src)
	}
	wg.Wait()
	return discovered
}

func (h *Handlers) mergeDiscoveredSources(entry domain.DictionaryEntry, discovered []struct {
	name string
	id   string
}) domain.DictionaryEntry {
	for _, ds := range discovered {
		if _, exists := entry.Sources[ds.name]; !exists {
			if entry.Sources == nil {
				entry.Sources = make(map[string]string)
			}
			entry.Sources[ds.name] = ds.id
		}
	}
	return entry
}

func (h *Handlers) refreshWithOldState(ctx context.Context, id string, oldSources map[string]string) (bool, error) {
	entry, found, err := h.Repo.GetDictionary(ctx, id)
	if err != nil || !found {
		return false, err
	}

	ordered := h.Registry.Ordered()
	prio := buildPriorityIndex(ordered)
	newStats := h.fetchAllSourceStats(ctx, entry.Sources, ordered)

	entry.SourceStats = newStats
	entry.BestSource = pickBestSource(newStats, prio)

	changed := hasEntryChanged(entry, oldSources)

	if entry.CoverURL == "" {
		entry.CoverURL = h.fetchCover(ctx, &entry, ordered)
	}

	now := time.Now()
	entry.UpdatedAt = &now
	if err := h.Repo.UpsertDictionary(ctx, entry); err != nil {
		return false, err
	}
	return changed, nil
}

func buildPriorityIndex(ordered []domain.Scraper) map[string]int {
	m := make(map[string]int, len(ordered))
	for i, src := range ordered {
		m[src.Source()] = i
	}
	return m
}

func (h *Handlers) fetchAllSourceStats(ctx context.Context, sources map[string]string, ordered []domain.Scraper) map[string]domain.SourceStat {
	srcMap := make(map[string]domain.Scraper, len(ordered))
	for _, s := range ordered {
		srcMap[s.Source()] = s
	}

	type result struct {
		name string
		stat domain.SourceStat
	}
	var g concurrent.Collect[result]
	for name, id := range sources {
		src, _ := srcMap[name]
		g.Go(func() result {
			return result{name: name, stat: h.fetchSourceStat(ctx, src, id)}
		})
	}

	stats := make(map[string]domain.SourceStat, len(sources))
	for _, r := range g.Wait() {
		stats[r.name] = r.stat
	}
	return stats
}

func hasEntryChanged(entry domain.DictionaryEntry, oldSources map[string]string) bool {
	if oldSources != nil {
		if len(oldSources) != len(entry.Sources) {
			return true
		}
	}
	return sourcesChanged(entry.Sources, oldSources)
}

func (h *Handlers) fetchSourceStat(ctx context.Context, src domain.Scraper, sourceID string) domain.SourceStat {
	stat := domain.SourceStat{ChaptersByLang: map[string]int{}, FetchedAt: time.Now()}
	if src == nil {
		stat.Err = "scraper not registered"
		return stat
	}
	chapters, err := src.FetchChapterList(ctx, sourceID)
	if err != nil {
		stat.Err = err.Error()
		return stat
	}
	for _, ch := range chapters {
		stat.ChaptersByLang[ch.Language]++
	}
	return stat
}

func (h *Handlers) fetchCover(ctx context.Context, entry *domain.DictionaryEntry, ordered []domain.Scraper) string {
	for _, src := range ordered {
		sourceID, ok := entry.Sources[src.Source()]
		if !ok {
			continue
		}
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil || manga == nil || manga.CoverURL == "" {
			continue
		}
		data, err := h.Downloader.Download(ctx, manga.CoverURL)
		if err != nil {
			continue
		}
		ext := extOrDefault(manga.CoverURL)
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "image/jpeg"
		}
		url, err := h.S3.Upload(ctx, entry.ID+"/cover"+ext, data, ct)
		if err != nil {
			h.Log.Warn("dictionary refresh: upload cover", "id", entry.ID, "err", err)
			continue
		}
		return url
	}
	return ""
}

func extOrDefault(url string) string {
	if ext := filepath.Ext(url); ext != "" {
		return ext
	}
	return ".jpg"
}

func (h *Handlers) buildEntries(hits map[string]*searchHit, existingIDs map[string]string) []domain.DictionaryEntry {
	entries := make([]domain.DictionaryEntry, 0, len(hits))
	now := time.Now()
	for _, hit := range hits {
		id := uuid.New().String()
		if existingID, ok := existingIDs[hit.slug]; ok {
			id = existingID
		}
		entries = append(entries, domain.DictionaryEntry{
			ID:          id,
			Slug:        hit.slug,
			Title:       hit.title,
			Sources:     hit.sources,
			SourceStats: hit.sourceStats,
			BestSource:  map[string]string{},
			CoverURL:    "",
			UpdatedAt:   &now,
			CreatedAt:   now,
		})
	}
	return entries
}

func (h *Handlers) logBatchUpsert(entries []domain.DictionaryEntry) {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	h.Log.Info("[DictionaryService] Search: batch upserted", "count", len(entries), "ids", ids)
}
