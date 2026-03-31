package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"manga-engine/config"
	"manga-engine/internal/domain"
	"manga-engine/pkg/concurrent"
)

const (
	defaultBatchSize          = 100
	defaultWorkerCount        = 5
	defaultCleanupWorkerCount = 5
)

type IngestConfig struct {
	Repo               domain.Repository
	DictSvc            domain.DictionaryHandler
	Cfg                *config.Config
	BatchSize          int
	WorkerCount        int
	CleanupWorkerCount int
}

func (c *IngestConfig) batchSize() int {
	if c.BatchSize <= 0 {
		return defaultBatchSize
	}
	return c.BatchSize
}

func (c *IngestConfig) workerCount() int {
	if c.WorkerCount <= 0 {
		return defaultWorkerCount
	}
	return c.WorkerCount
}

func (c *IngestConfig) cleanupWorkerCount() int {
	if c.CleanupWorkerCount <= 0 {
		return defaultCleanupWorkerCount
	}
	return c.CleanupWorkerCount
}

type IngestDaemon struct {
	repo     domain.Repository
	dict     domain.DictionaryHandler
	diskPath string
	interval time.Duration
	log      *slog.Logger
	cfg      IngestConfig
}

func NewIngestDaemon(cfg IngestConfig) *IngestDaemon {
	return &IngestDaemon{
		repo:     cfg.Repo,
		dict:     cfg.DictSvc,
		diskPath: cfg.Cfg.LibraryPath,
		interval: cfg.Cfg.DictionaryRefreshInterval,
		log:      slog.With("daemon", "ingest"),
		cfg:      cfg,
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

func (d *IngestDaemon) refreshMangaEntries(ctx context.Context) {
	pool := concurrent.NewWorkerPool[string](d.cfg.workerCount(), d.cfg.workerCount()*2)
	pool.Start(ctx, func(ctx context.Context, dictID string) {
		if _, err := d.dict.Refresh(ctx, dictID); err != nil {
			d.log.Warn("ingest daemon: refresh", "dictionaryID", dictID, "err", err)
		}
	})
	defer pool.Wait()

	filter := domain.MangaFilter{
		PageSize: d.cfg.batchSize(),
		States:   []string{string(domain.StateAvailable), string(domain.StateFetching)},
	}
	for page := 1; ; page++ {
		if ctx.Err() != nil {
			return
		}
		filter.Page = page
		result, err := d.repo.ListManga(ctx, filter)
		if err != nil {
			d.log.Error("ingest daemon: list manga", "err", err)
			return
		}
		if len(result.Items) == 0 {
			break
		}
		for _, m := range result.Items {
			select {
			case pool.SubmitChan() <- m.DictionaryID:
			case <-ctx.Done():
				return
			}
		}
		if len(result.Items) < d.cfg.batchSize() {
			break
		}
	}
}

func (d *IngestDaemon) cleanupOrphanedDirs(ctx context.Context) {
	mangaIDSet := make(map[string]bool)
	var mu sync.Mutex

	pool := concurrent.NewWorkerPool[string](d.cfg.cleanupWorkerCount(), d.cfg.cleanupWorkerCount()*2)
	pool.Start(ctx, func(ctx context.Context, id string) {
		mu.Lock()
		mangaIDSet[id] = true
		mu.Unlock()
	})
	defer pool.Wait()

	filter := domain.MangaFilter{PageSize: d.cfg.batchSize()}
	for page := 1; ; page++ {
		if ctx.Err() != nil {
			return
		}
		filter.Page = page
		result, err := d.repo.ListManga(ctx, filter)
		if err != nil {
			d.log.Error("ingest daemon: list manga for cleanup", "err", err)
			return
		}
		if len(result.Items) == 0 {
			break
		}
		for _, m := range result.Items {
			select {
			case pool.SubmitChan() <- m.ID:
			case <-ctx.Done():
				return
			}
		}
		if len(result.Items) < d.cfg.batchSize() {
			break
		}
	}

	entries, err := os.ReadDir(d.diskPath)
	if err != nil {
		d.log.Error("ingest daemon: read library dir", "err", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mu.Lock()
		exists := mangaIDSet[e.Name()]
		mu.Unlock()
		if !exists {
			dirPath := filepath.Join(d.diskPath, e.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				d.log.Warn("ingest daemon: remove orphaned dir", "dir", e.Name(), "err", err)
			} else {
				d.log.Info("ingest daemon: removed orphaned dir", "dir", e.Name())
			}
		}
	}
}
