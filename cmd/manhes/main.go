// @title           Manhes API
// @version         2.0
// @description     Manga ingestion and catalog API.
// @basePath        /api/v1
// @accept          json
// @produce         json

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "manga-engine/docs/manhes"

	"manga-engine/config"
	"manga-engine/internal/application"
	"manga-engine/internal/infrastructure/downloader"
	"manga-engine/internal/handler"
	"manga-engine/internal/infrastructure/messaging"
	"manga-engine/internal/infrastructure/persistence"
	"manga-engine/internal/infrastructure/s3"
	"manga-engine/internal/infrastructure/scraper"
	"manga-engine/internal/infrastructure/scraper/atsu"
	"manga-engine/internal/infrastructure/scraper/mangadex"
	"manga-engine/internal/infrastructure/storage"
	"manga-engine/pkg/lifecycle"
	pkglog "manga-engine/pkg/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Default().Error("[Config] failed to load", "err", err)
		os.Exit(1)
	}
	log := pkglog.New(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("[Core] server stopped", "err", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config, log *slog.Logger) error {
	ctx, stop := lifecycle.WithShutdown(context.Background())
	defer stop()

	log.Info("[Core] starting", "addr", cfg.ListenAddr, "log_level", cfg.LogLevel)

	repo, err := persistence.NewSQLite(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer repo.Close()
	log.Info("[DB] connected", "path", cfg.DBPath)

	s3c, err := s3.New(ctx, cfg.S3)
	if err != nil {
		return fmt.Errorf("s3: %w", err)
	}
	log.Info("[S3] connected", "endpoint", cfg.S3.Endpoint, "bucket", cfg.S3.Bucket)

	disk := storage.NewDisk(cfg.LibraryPath)

	producer := messaging.NewProducer(cfg)
	defer producer.Close()
	log.Info("[Kafka] producer ready", "brokers", cfg.Kafka.Brokers)

	reg := buildScraperRegistry(cfg)
	dl := downloader.New(&http.Client{Timeout: cfg.DownloaderTimeout})

	ingestSvc := application.NewIngestService(repo, reg, dl, disk, producer)
	syncSvc := application.NewSyncService(repo, disk, s3c, application.SyncConfig{
		LibraryPath: cfg.LibraryPath,
		Interval:    cfg.SyncInterval,
	})
	dictSvc := application.NewDictionaryService(repo, reg, dl, s3c, application.DictionaryConfig{
		RefreshInterval: cfg.DictionaryRefreshInterval,
	})
	watchlistSvc := application.NewWatchlistService(repo, dictSvc, reg, producer)
	catalogSvc := application.NewCatalogService(repo)

	ingestConsumer := messaging.NewIngestConsumer(cfg, ingestSvc)
	syncConsumer := messaging.NewSyncConsumer(cfg, syncSvc)

	go ingestConsumer.Run(ctx)
	go syncConsumer.Run(ctx)
	log.Info("[Kafka] consumers started")

	go syncSvc.SyncAll(ctx)
	go watchlistSvc.RunDaemon(ctx)
	go dictSvc.RunDaemon(ctx)
	log.Info("[Watchlist] daemon started")
	log.Info("[Dictionary] daemon started", "interval", cfg.DictionaryRefreshInterval)

	log.Info("[Core] server up", "addr", cfg.ListenAddr)
	h := handler.NewHandlers(watchlistSvc, catalogSvc, dictSvc, log)
	return handler.NewServer(cfg.ListenAddr, handler.NewRouter(h, cfg), log).Run(ctx)
}

func buildScraperRegistry(cfg *config.Config) *scraper.Registry {
	reg := &scraper.Registry{}
	client := &http.Client{Timeout: cfg.DownloaderTimeout}

	if cfg.Mangadex.RateLimit > 0 {
		reg.Register(1, mangadex.New(cfg.Mangadex, client))
	}
	if cfg.Atsu.RateLimit > 0 {
		reg.Register(2, atsu.New(cfg.Atsu, client))
	}

	return reg
}
