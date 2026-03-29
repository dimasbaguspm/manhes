package main

import (
	"context"
	"log/slog"
	"net/http"

	"manga-engine/config"
	"manga-engine/internal/application"
	"manga-engine/internal/daemon"
	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/downloader"
	"manga-engine/internal/infrastructure/eventbus"
	"manga-engine/internal/infrastructure/persistence"
	"manga-engine/internal/infrastructure/s3"
	"manga-engine/internal/infrastructure/scraper"
	"manga-engine/internal/infrastructure/scraper/atsu"
	"manga-engine/internal/infrastructure/scraper/mangadex"
	"manga-engine/internal/infrastructure/storage"
	"manga-engine/internal/subscriber"
)

type Infra struct {
	Repo domain.Repository
	S3   domain.ObjectStore
	Disk domain.Storer
	Bus  domain.EventBus
	Reg  *scraper.Registry
	DL   domain.Downloader
	Cfg  *config.Config
	Log  *slog.Logger
}

type Wiring struct {
	Infra Infra

	DictSvc       *application.DictionaryService
	MangaSvc      *application.MangaService
	RetrievalSub  *subscriber.RetrievalSubscriber
	MangaSub      *subscriber.MangaSubscriber
	DictionarySub *subscriber.DictionarySubscriber
	FileUploadSub *subscriber.FileUploadSubscriber
	IngestDaemon  *daemon.IngestDaemon
}

func New(infra Infra) *Wiring {
	dictSvc := application.NewDictionaryService(application.DictionaryServiceConfig{
		Repo:     infra.Repo,
		Registry: infra.Reg,
		DL:       infra.DL,
		S3:       infra.S3,
		Bus:      infra.Bus,
		Cfg:      infra.Cfg,
	})

	mangaSvc := application.NewMangaService(application.MangaServiceConfig{
		Repo: infra.Repo,
		Cfg:  infra.Cfg,
	})

	retrievalSub := subscriber.NewRetrievalSubscriber(subscriber.RetrievalSubscriberConfig{
		Repo:     infra.Repo,
		Registry: infra.Reg,
		Bus:      infra.Bus,
		Cfg:      infra.Cfg,
	})

	mangaSub := subscriber.NewMangaSubscriber(subscriber.MangaSubscriberConfig{
		Repo:     infra.Repo,
		Registry: infra.Reg,
		Bus:      infra.Bus,
		Cfg:      infra.Cfg,
	})

	dictionarySub := subscriber.NewDictionarySubscriber(subscriber.DictionarySubscriberConfig{
		DictionaryManager: dictSvc,
		Cfg:               infra.Cfg,
	})

	fileUploadSub := subscriber.NewFileUploadSubscriber(subscriber.FileUploadSubscriberConfig{
		Repo:     infra.Repo,
		Registry: infra.Reg,
		Disk:     infra.Disk,
		S3:       infra.S3,
		DL:       infra.DL,
		Bus:      infra.Bus,
		Cfg:      infra.Cfg,
	})

	ingestDaemon := daemon.NewIngestDaemon(daemon.IngestConfig{
		Repo:    infra.Repo,
		DictSvc: dictSvc,
		Cfg:     infra.Cfg,
	})

	return &Wiring{
		Infra:         infra,
		DictSvc:       dictSvc,
		MangaSvc:      mangaSvc,
		RetrievalSub:  retrievalSub,
		MangaSub:      mangaSub,
		DictionarySub: dictionarySub,
		FileUploadSub: fileUploadSub,
		IngestDaemon:  ingestDaemon,
	}
}

func (w *Wiring) StartSubscriptions() {
	log := w.Infra.Log
	bus := w.Infra.Bus
	cfg := w.Infra.Cfg

	log.Info("[Wiring] registering event subscriptions")
	bus.Subscribe(cfg.Bus.DictionaryUpdated, func(ctx context.Context, e domain.Event) error {
		return w.MangaSub.HandleDictionaryUpdated(ctx, e.(domain.DictionaryUpdated))
	})
	bus.Subscribe(cfg.Bus.DictionaryRefreshed, func(ctx context.Context, e domain.Event) error {
		return w.DictionarySub.HandleDictionaryRefreshed(ctx, e.(domain.DictionaryRefreshed))
	})
	bus.Subscribe(cfg.Bus.IngestRequested, func(ctx context.Context, e domain.Event) error {
		return w.RetrievalSub.HandleIngestRequested(ctx, e.(domain.IngestRequested))
	})
	bus.Subscribe(cfg.Bus.ChaptersFound, func(ctx context.Context, e domain.Event) error {
		return w.FileUploadSub.HandleChaptersFound(ctx, e.(domain.ChaptersFound))
	})
	bus.Subscribe(cfg.Bus.ChapterUploaded, func(ctx context.Context, e domain.Event) error {
		return w.MangaSub.HandleChapterUploaded(ctx, e.(domain.ChapterUploaded))
	})
	bus.Subscribe(cfg.Bus.MangaAvailable, func(ctx context.Context, e domain.Event) error {
		return w.MangaSub.HandleMangaAvailable(ctx, e.(domain.MangaAvailable))
	})
	log.Info("[Wiring] event subscriptions registered")
}

func (w *Wiring) StartDaemons(ctx context.Context) {
	log := w.Infra.Log

	log.Info("[Wiring] starting daemons")
	go w.IngestDaemon.Run(ctx)
	log.Info("[Wiring] daemons started")
}

func BuildScraperRegistry(cfg *config.Config) *scraper.Registry {
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

func InitInfra(ctx context.Context, cfg *config.Config, log *slog.Logger) (Infra, error) {
	repo, err := persistence.NewMySQL(ctx, cfg.Database)
	if err != nil {
		return Infra{}, err
	}
	log.Info("[DB] connected", "host", cfg.Database.Host, "port", cfg.Database.Port, "db", cfg.Database.Name)

	s3c, err := s3.New(ctx, cfg.S3)
	if err != nil {
		repo.Close()
		return Infra{}, err
	}
	log.Info("[S3] connected", "endpoint", cfg.S3.Endpoint, "bucket", cfg.S3.Bucket)

	disk := storage.NewDisk(cfg.LibraryPath)
	bus := eventbus.New()
	reg := BuildScraperRegistry(cfg)
	dl := downloader.New(&http.Client{Timeout: cfg.DownloaderTimeout})

	return Infra{
		Repo: repo,
		S3:   s3c,
		Disk: disk,
		Bus:  bus,
		Reg:  reg,
		DL:   dl,
		Cfg:  cfg,
		Log:  log,
	}, nil
}
