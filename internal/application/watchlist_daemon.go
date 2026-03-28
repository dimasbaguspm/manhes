package application

import (
	"context"
	"time"

	"manga-engine/internal/domain"
)

const retryInterval = time.Hour

func (s *WatchlistService) RunDaemon(ctx context.Context) {
	s.recoverStuck(ctx)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkDue(ctx)
		}
	}
}

// recoverStuck resets the cooldown for any watchlist entries whose dictionary is
// not yet available (fetching/uploading/unavailable), then immediately triggers
// checkDue so they are re-ingested instead of waiting up to retryInterval.
func (s *WatchlistService) recoverStuck(ctx context.Context) {
	entries, err := s.repo.ListWatchlist(ctx)
	if err != nil {
		s.log.Error("watchlist daemon: recover stuck: list watchlist", "err", err)
		return
	}
	recovered := 0
	for _, e := range entries {
		if e.DictionaryID == "" {
			continue
		}
		dictEntry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
		if err != nil || !found {
			continue
		}
		if dictEntry.State == domain.StateAvailable {
			continue
		}
		s.log.Info("watchlist daemon: resetting stuck entry on startup",
			"slug", e.Slug,
			"state", string(dictEntry.State),
		)
		if err := s.repo.UpdateLastChecked(ctx, e.Slug, time.Time{}); err != nil {
			s.log.Warn("watchlist daemon: reset last checked", "slug", e.Slug, "err", err)
		}
		recovered++
	}
	if recovered == 0 {
		s.log.Info("watchlist daemon: no stuck entries on startup")
	}
	s.checkDue(ctx)
}

func (s *WatchlistService) checkDue(ctx context.Context) {
	entries, err := s.repo.ListWatchlist(ctx)
	if err != nil {
		s.log.Error("watchlist daemon: list watchlist", "err", err)
		return
	}
	now := time.Now()
	for _, e := range entries {
		// If the dictionary is now available, the ingest is fulfilled — clean up.
		if e.DictionaryID != "" {
			dictEntry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
			if err == nil && found && dictEntry.State == domain.StateAvailable {
				s.log.Info("watchlist daemon: fulfilled, removing", "slug", e.Slug)
				if err := s.repo.RemoveWatchlist(ctx, e.Slug); err != nil {
					s.log.Warn("watchlist daemon: remove fulfilled entry", "slug", e.Slug, "err", err)
				}
				continue
			}
		}

		// Cooldown: don't re-trigger if we checked recently.
		if e.LastChecked != nil && now.Sub(*e.LastChecked) < retryInterval {
			continue
		}

		sources, langToSource, err := s.resolveSourcesForEntry(ctx, &e)
		if err != nil {
			s.log.Error("watchlist daemon: resolve sources", "slug", e.Slug, "err", err)
			continue
		}
		s.log.Info("watchlist daemon: triggering ingest", "slug", e.Slug)
		if err := s.publisher.PublishIngestRequested(ctx, domain.IngestRequested{
			Slug:         e.Slug,
			Sources:      sources,
			LangToSource: langToSource,
		}); err != nil {
			s.log.Error("watchlist daemon: publish", "slug", e.Slug, "err", err)
			continue
		}
		if e.DictionaryID != "" {
			if err := s.repo.SetDictionaryState(ctx, e.DictionaryID, domain.StateFetching); err != nil {
				s.log.Warn("watchlist daemon: set dictionary state fetching", "slug", e.Slug, "err", err)
			}
		}
		if err := s.repo.UpdateLastChecked(ctx, e.Slug, now); err != nil {
			s.log.Warn("watchlist daemon: update last checked", "slug", e.Slug, "err", err)
		}
	}
}
