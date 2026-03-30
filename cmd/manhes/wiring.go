package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"manga-engine/config"
	"manga-engine/internal/daemon"
	"manga-engine/internal/domain"
	"manga-engine/internal/handler"
	"manga-engine/internal/infrastructure/downloader"
	"manga-engine/internal/infrastructure/eventbus"
	infrahttp "manga-engine/internal/infrastructure/http"
	"manga-engine/internal/infrastructure/persistence"
	"manga-engine/internal/ui"

	"manga-engine/internal/infrastructure/s3"
	"manga-engine/internal/infrastructure/scraper"
	"manga-engine/internal/infrastructure/scraper/atsu"
	"manga-engine/internal/infrastructure/scraper/mangadex"
	"manga-engine/internal/infrastructure/storage"
	"manga-engine/internal/subscriber"

	httpSwagger "github.com/swaggo/http-swagger"
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

	*handler.Handlers

	RetrievalSub  *subscriber.RetrievalSubscriber
	MangaSub      *subscriber.MangaSubscriber
	DictionarySub *subscriber.DictionarySubscriber
	FileUploadSub *subscriber.FileUploadSubscriber
	IngestDaemon  *daemon.IngestDaemon
}

func New(infra Infra) *Wiring {
	h := handler.NewHandlers(handler.HandlersConfig{
		Repo:       infra.Repo,
		Registry:   infra.Reg,
		Downloader: infra.DL,
		S3:         infra.S3,
		Bus:        infra.Bus,
		Cfg:        infra.Cfg,
		Log:        infra.Log,
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
		DictionaryHandler: h,
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
		DictSvc: h, // *Handlers implements DictionaryManager
		Cfg:     infra.Cfg,
	})

	return &Wiring{
		Infra:         infra,
		Handlers:      h,
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

	log.Info("[Core] registering event subscriptions")
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
	log.Info("[Core] event subscriptions registered")
}

func (w *Wiring) StartDaemons(ctx context.Context) {
	log := w.Infra.Log

	log.Info("[Core] starting daemons 2")
	go w.IngestDaemon.Run(ctx)
	log.Info("[Core] daemons started")
}

func (w *Wiring) RegisterRoutes() {
	w.Infra.Log.Info("[Core] registering HTTP routes")
}

func (w *Wiring) StartServer(ctx context.Context) error {
	log := w.Infra.Log

	router := infrahttp.NewMux()
	router.Use(infrahttp.Cors)
	router.Use(infrahttp.RequestID)
	router.Use(infrahttp.StructuredLogger(w.Handlers.Log))
	router.Use(middleware.Recoverer)

	router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	h := &httpHandlers{Handlers: w.Handlers}

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/manga", h.listManga)
		r.Get("/manga/{mangaId}", h.getManga)
		r.Get("/manga/{mangaId}/{lang}", h.getChapters)
		r.Get("/read/{chapterId}", h.readChapter)
		r.Get("/dictionary", h.searchDictionary)
		r.Post("/dictionary/refresh", h.refreshDictionary)
	})

	router.Handle("/*", ui.NewHandler())

	srv := &http.Server{
		Addr:         w.Infra.Cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("[Core] server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	}
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
