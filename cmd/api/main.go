package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/app"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	log := observability.NewLogger(cfg.App.Environment, cfg.Observability.LogLevel)
	log.Info("starting application")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := observability.InitTracing(ctx, cfg.App.Name, cfg.Observability.OTELExporterEndpoint, cfg.Observability.OTELEnabled)
	if err != nil {
		log.Error("tracing startup failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownTracing(context.Background()) }()

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("startup failed", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			stop()
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", "error", err)
	}
}
