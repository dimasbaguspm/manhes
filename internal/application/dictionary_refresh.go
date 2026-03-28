package application

import (
	"context"
	"mime"
	"path/filepath"
	"time"

	"manga-engine/internal/domain"
	"manga-engine/pkg/concurrent"
)

// RefreshAll refreshes all dictionary entries and publishes DictionaryUpdated events
// for entries that have changed.
func (s *DictionaryService) RefreshAll(ctx context.Context) {
	s.refreshAll(ctx)
}

func (s *DictionaryService) refreshAll(ctx context.Context) {
	// Fetch all entries in one large page (no UI pagination needed here).
	page, err := s.repo.ListDictionary(ctx, domain.DictionaryFilter{PageSize: 10_000})
	if err != nil {
		s.log.Error("dictionary daemon: list", "err", err)
		return
	}
	for _, e := range page.Items {
		if ctx.Err() != nil {
			return
		}
		if _, err := s.refresh(ctx, e.ID); err != nil {
			s.log.Warn("dictionary daemon: refresh", "id", e.ID, "slug", e.Slug, "err", err)
		}
	}
}

func (s *DictionaryService) refresh(ctx context.Context, id string) (bool, error) {
	return s.refreshWithOldState(ctx, id, nil)
}

// refreshWithOldState fetches source stats for all known sources, optionally
// compares against pre-merge oldSources to detect changes, and persists updates.
func (s *DictionaryService) refreshWithOldState(ctx context.Context, id string, oldSources map[string]string) (bool, error) {
	entry, found, err := s.repo.GetDictionary(ctx, id)
	if err != nil || !found {
		return false, err
	}

	ordered := s.registry.Ordered()

	// Build priority index for tie-breaking when two sources have equal chapter counts.
	priorityIndex := make(map[string]int, len(ordered))
	for i, src := range ordered {
		priorityIndex[src.Source()] = i
	}

	type sourceStatResult struct {
		name string
		stat domain.SourceStat
	}

	var g concurrent.Collect[sourceStatResult]
	for sourceName, sourceID := range entry.Sources {
		var src domain.Scraper
		for _, candidate := range ordered {
			if candidate.Source() == sourceName {
				src = candidate
				break
			}
		}
		g.Go(func() sourceStatResult {
			return sourceStatResult{name: sourceName, stat: s.fetchSourceStat(ctx, src, sourceID)}
		})
	}

	type best struct {
		source string
		count  int
		prio   int
	}
	newStats := make(map[string]domain.SourceStat, len(entry.Sources))
	// langBest tracks the highest-chapter-count source per language.
	langBest := make(map[string]best)
	for _, r := range g.Wait() {
		newStats[r.name] = r.stat
		prio := priorityIndex[r.name]
		for lang, count := range r.stat.ChaptersByLang {
			if b, seen := langBest[lang]; !seen || count > b.count || (count == b.count && prio < b.prio) {
				langBest[lang] = best{source: r.name, count: count, prio: prio}
			}
		}
	}

	newBest := make(map[string]string, len(langBest))
	for lang, b := range langBest {
		newBest[lang] = b.source
	}

	// Check if sources, stats, or best-source mapping changed.
	changed := entryChanged(entry, oldSources, newStats, newBest)

	if entry.CoverURL == "" {
		entry.CoverURL = s.fetchCover(ctx, entry, ordered)
	}

	now := time.Now()
	entry.SourceStats = newStats
	entry.BestSource = newBest
	entry.UpdatedAt = &now
	if err := s.repo.UpsertDictionary(ctx, entry); err != nil {
		return false, err
	}

	return changed, nil
}

// entryChanged compares pre-merge (oldSources) and post-merge (entry.Sources) state
// along with SourceStats and BestSource to determine if the entry meaningfully changed.
func entryChanged(entry domain.DictionaryEntry, oldSources map[string]string, newStats map[string]domain.SourceStat, newBest map[string]string) bool {
	// Compare Sources: check if any source was added or removed.
	if oldSources != nil {
		if len(oldSources) != len(entry.Sources) {
			return true
		}
		for name, oldID := range oldSources {
			if entry.Sources[name] != oldID {
				return true
			}
		}
	}

	// Compare SourceStats: check if total chapter counts per language changed.
	if statsChanged(entry.SourceStats, newStats) {
		return true
	}

	// Compare BestSource: check if language→source mapping changed.
	if len(entry.BestSource) != len(newBest) {
		return true
	}
	for lang, oldSrc := range entry.BestSource {
		if newSrc, ok := newBest[lang]; !ok || newSrc != oldSrc {
			return true
		}
	}

	return false
}

// statsChanged compares two SourceStats maps and returns true if the total
// chapter counts per language or the set of available languages differ.
func statsChanged(oldStats, newStats map[string]domain.SourceStat) bool {
	// Collect total chapters per language from old and new stats.
	oldLangs := make(map[string]int)
	newLangs := make(map[string]int)

	for _, stat := range oldStats {
		for lang, count := range stat.ChaptersByLang {
			oldLangs[lang] += count
		}
	}
	for _, stat := range newStats {
		for lang, count := range stat.ChaptersByLang {
			newLangs[lang] += count
		}
	}

	// Compare language sets and chapter counts.
	if len(oldLangs) != len(newLangs) {
		return true
	}
	for lang, oldCount := range oldLangs {
		if newCount, ok := newLangs[lang]; !ok || newCount != oldCount {
			return true
		}
	}
	return false
}

func (s *DictionaryService) fetchSourceStat(ctx context.Context, src domain.Scraper, sourceID string) domain.SourceStat {
	stat := domain.SourceStat{
		ChaptersByLang: map[string]int{},
		FetchedAt:      time.Now(),
	}
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

func (s *DictionaryService) fetchCover(ctx context.Context, entry domain.DictionaryEntry, ordered []domain.Scraper) string {
	for _, src := range ordered {
		sourceID, ok := entry.Sources[src.Source()]
		if !ok {
			continue
		}
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil || manga == nil || manga.CoverURL == "" {
			continue
		}
		data, err := s.dl.Download(ctx, manga.CoverURL)
		if err != nil {
			continue
		}
		ext := filepath.Ext(manga.CoverURL)
		if ext == "" {
			ext = ".jpg"
		}
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "image/jpeg"
		}
		url, err := s.s3c.Upload(ctx, entry.ID+"/cover"+ext, data, ct)
		if err != nil {
			s.log.Warn("dictionary refresh: upload cover", "id", entry.ID, "err", err)
			continue
		}
		return url
	}
	return ""
}
