package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

// IngestConfig holds dependencies for the ingest daemon.
type IngestConfig struct {
	Repo    domain.Repository
	DictSvc domain.DictionaryHandler
	Cfg     *config.Config
}

// IngestDaemon periodically refreshes manga dictionary entries and cleans up
// orphaned manga directories on disk.
type IngestDaemon struct {
	repo     domain.Repository
	dict     domain.DictionaryHandler
	diskPath string
	interval time.Duration
	log      *slog.Logger
}

// NewIngestDaemon creates a new IngestDaemon.
func NewIngestDaemon(cfg IngestConfig) *IngestDaemon {
	return &IngestDaemon{
		repo:     cfg.Repo,
		dict:     cfg.DictSvc,
		diskPath: cfg.Cfg.LibraryPath,
		interval: cfg.Cfg.DictionaryRefreshInterval,
		log:      slog.With("daemon", "ingest"),
	}
}

// Run starts the periodic refresh and cleanup loop.
func (d *IngestDaemon) Run(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.refreshMangaEntries(ctx)
			d.cleanupOrphanedDirs(ctx)
		}
	}
}

// refreshMangaEntries lists manga with state "available" or "fetching" and
// refreshes their corresponding dictionary entries.
func (d *IngestDaemon) refreshMangaEntries(ctx context.Context) {
	page, err := d.repo.ListManga(ctx, domain.MangaFilter{
		PageSize: 10_000,
		States:   []string{string(domain.StateAvailable), string(domain.StateFetching)},
	})
	if err != nil {
		d.log.Error("ingest daemon: list manga", "err", err)
		return
	}
	for _, m := range page.Items {
		if ctx.Err() != nil {
			return
		}
		if _, err := d.dict.Refresh(ctx, m.DictionaryID); err != nil {
			d.log.Warn("ingest daemon: refresh", "dictionaryID", m.DictionaryID, "err", err)
		}
	}
}

// cleanupOrphanedDirs scans the disk library directory and removes any manga
// directories that no longer have a corresponding entry in the manga table.
func (d *IngestDaemon) cleanupOrphanedDirs(ctx context.Context) {
	// Build set of manga IDs from DB.
	page, err := d.repo.ListManga(ctx, domain.MangaFilter{PageSize: 10_000})
	if err != nil {
		d.log.Error("ingest daemon: list manga for cleanup", "err", err)
		return
	}
	mangaIDSet := make(map[string]bool)
	for _, m := range page.Items {
		mangaIDSet[m.ID] = true
	}

	// Scan disk and remove orphaned dirs.
	entries, err := os.ReadDir(d.diskPath)
	if err != nil {
		d.log.Error("ingest daemon: read library dir", "err", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !mangaIDSet[e.Name()] {
			dirPath := filepath.Join(d.diskPath, e.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				d.log.Warn("ingest daemon: remove orphaned dir", "dir", e.Name(), "err", err)
			} else {
				d.log.Info("ingest daemon: removed orphaned dir", "dir", e.Name())
			}
		}
	}
}
