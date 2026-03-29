// @title           Manhes API
// @version         2.0
// @description     Manga ingestion and catalog API.
// @basePath        /api/v1
// @accept          json
// @produce         json

package main

import (
	"context"
	"log/slog"
	"os"

	_ "manga-engine/docs/manhes"

	"manga-engine/config"
	"manga-engine/internal/handler"
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

	ctx, stop := lifecycle.WithShutdown(context.Background())
	defer stop()

	log.Info("[Core] starting", "addr", cfg.ListenAddr, "log_level", cfg.LogLevel)

	infra, err := InitInfra(ctx, cfg, log)
	if err != nil {
		log.Error("[Core] infra init failed", "err", err)
		os.Exit(1)
	}
	defer infra.Repo.Close()

	w := New(infra)
	w.StartSubscriptions()
	w.StartDaemons(ctx)

	log.Info("[Core] server up", "addr", cfg.ListenAddr)
	h := handler.NewHandlers(w.MangaSvc, w.DictSvc, infra.Bus, cfg, log)
	if err := handler.NewServer(cfg.ListenAddr, handler.NewRouter(h, cfg), log).Run(ctx); err != nil {
		log.Error("[Core] server stopped", "err", err)
		os.Exit(1)
	}
}
