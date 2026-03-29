package handler

import (
	"log/slog"
	"net/http"

	"manga-engine/config"
	"manga-engine/internal/domain"
	"manga-engine/pkg/httputil"
	"manga-engine/pkg/reqctx"
)

type Handlers struct {
	Repo       domain.Repository
	Registry   domain.SourceRegistry
	Downloader domain.Downloader
	S3         domain.ObjectStore
	Bus        domain.EventBus
	Cfg        *config.Config
	Log        *slog.Logger
}

var _ domain.DictionaryHandler = (*Handlers)(nil)
var _ domain.MangaHandler = (*Handlers)(nil)
var _ domain.ChapterHandler = (*Handlers)(nil)

type HandlersConfig struct {
	Repo       domain.Repository
	Registry   domain.SourceRegistry
	Downloader domain.Downloader
	S3         domain.ObjectStore
	Bus        domain.EventBus
	Cfg        *config.Config
	Log        *slog.Logger
}

func NewHandlers(cfg HandlersConfig) *Handlers {
	return &Handlers{
		Repo:       cfg.Repo,
		Registry:   cfg.Registry,
		Downloader: cfg.Downloader,
		S3:         cfg.S3,
		Bus:        cfg.Bus,
		Cfg:        cfg.Cfg,
		Log:        cfg.Log,
	}
}

func (h *Handlers) internalError(w http.ResponseWriter, r *http.Request, op string, err error) {
	h.Log.Error(op,
		slog.String("request_id", reqctx.RequestID(r.Context())),
		slog.Any("err", err),
	)
	httputil.InternalError(w, err)
}
