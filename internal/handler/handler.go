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
	manga      domain.MangaQuerier
	dictionary domain.DictionaryManager
	bus        domain.EventBus
	cfg        *config.Config
	log        *slog.Logger
}

func NewHandlers(
	manga domain.MangaQuerier,
	dictionary domain.DictionaryManager,
	bus domain.EventBus,
	cfg *config.Config,
	log *slog.Logger,
) *Handlers {
	return &Handlers{
		manga:      manga,
		dictionary: dictionary,
		bus:        bus,
		cfg:        cfg,
		log:        log,
	}
}

func (h *Handlers) internalError(w http.ResponseWriter, r *http.Request, op string, err error) {
	h.log.Error(op,
		slog.String("request_id", reqctx.RequestID(r.Context())),
		slog.Any("err", err),
	)
	httputil.InternalError(w, err)
}
