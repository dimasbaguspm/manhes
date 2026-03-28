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
	"manga-engine/internal/domain"
	"manga-engine/internal/handler"
	"manga-engine/internal/infrastructure/downloader"
	"manga-engine/internal/infrastructure/eventbus"
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

	repo, err := persistence.NewMySQL(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer repo.Close()
	log.Info("[DB] connected", "host", cfg.Database.Host, "port", cfg.Database.Port, "db", cfg.Database.Name)

	s3c, err := s3.New(ctx, cfg.S3)
	if err != nil {
		return fmt.Errorf("s3: %w", err)
	}
	log.Info("[S3] connected", "endpoint", cfg.S3.Endpoint, "bucket", cfg.S3.Bucket)

	disk := storage.NewDisk(cfg.LibraryPath)

	bus := eventbus.New()

	reg := buildScraperRegistry(cfg)
	dl := downloader.New(&http.Client{Timeout: cfg.DownloaderTimeout})

	dictSvc := application.NewDictionaryService(application.DictionaryServiceConfig{
		Repo:     repo,
		Registry: reg,
		DL:       dl,
		S3:       s3c,
		Bus:      bus,
		Cfg:      cfg,
	})
	retrievalHandler := application.NewRetrievalHandler(application.RetrievalHandlerConfig{
		Repo:     repo,
		Registry: reg,
		Bus:      bus,
		Cfg:      cfg,
	})
	mangaSvc := application.NewMangaService(application.MangaServiceConfig{
		Repo:     repo,
		Registry: reg,
		Bus:      bus,
		Cfg:      cfg,
	})
	fileUploadSvc := application.NewFileUploadService(application.FileUploadServiceConfig{
		Repo: repo,
		Disk: disk,
		S3:   s3c,
		DL:   dl,
		Bus:  bus,
		Cfg:  cfg,
	})
	ingestDaemon := application.NewIngestDaemon(application.IngestDaemonConfig{
		Repo:     repo,
		DictSvc:  dictSvc,
		DiskPath: cfg.LibraryPath,
		Cfg:      cfg,
	})

	// Subscribe event handlers
	bus.Subscribe(cfg.Bus.DictionaryUpdated, func(ctx context.Context, e domain.Event) error {
		return mangaSvc.HandleDictionaryUpdated(ctx, e.(domain.DictionaryUpdated))
	})
	bus.Subscribe(cfg.Bus.IngestRequested, func(ctx context.Context, e domain.Event) error {
		return retrievalHandler.HandleIngestRequested(ctx, e.(domain.IngestRequested))
	})
	bus.Subscribe(cfg.Bus.ChaptersFound, func(ctx context.Context, e domain.Event) error {
		return fileUploadSvc.HandleChaptersFound(ctx, e.(domain.ChaptersFound))
	})

	go ingestDaemon.Run(ctx)
	log.Info("[Ingest] daemon started", "interval", cfg.DictionaryRefreshInterval)

	log.Info("[Core] server up", "addr", cfg.ListenAddr)
	h := handler.NewHandlers(mangaSvc, dictSvc, log)
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
