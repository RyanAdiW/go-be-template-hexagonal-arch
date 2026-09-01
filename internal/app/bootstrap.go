package app

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/database"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/httpserver"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/observability"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/security"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user"
	"gopkg.in/go-playground/validator.v9"
	"log/slog"
	"net/http"
)

type Application struct {
	server  *http.Server
	closeDB func() error
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*Application, error) {
	db, closeDB, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}
	token := security.NewTokenManager(cfg.JWT)
	metrics := observability.NewHTTPMetrics(prometheus.DefaultRegisterer)
	router := httpserver.New(cfg, func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}, token, metrics, log)
	module := user.NewModule(db, validator.New())
	api := router.Group("/api/v1")
	module.RegisterRoutes(api, httpserver.Auth(token))
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), Handler: router, ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout, IdleTimeout: cfg.Server.IdleTimeout}
	return &Application{server, closeDB}, nil
}
func (a *Application) Run() error { return a.server.ListenAndServe() }
func (a *Application) Shutdown(ctx context.Context) error {
	err := a.server.Shutdown(ctx)
	closeErr := a.closeDB()
	if err != nil {
		return err
	}
	return closeErr
}
