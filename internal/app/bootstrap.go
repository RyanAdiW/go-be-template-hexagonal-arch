package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/database"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/httpserver"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/observability"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/security"
	"gopkg.in/go-playground/validator.v9"
	"gorm.io/gorm"
)

type Application struct {
	server  *http.Server
	closeDB func() error
}

type Dependencies struct {
	Now func() time.Time
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger, deps Dependencies) (*Application, error) {
	if deps.Now == nil {
		deps.Now = time.Now
	}

	db, closeDB, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	db = db.Session(&gorm.Session{NowFunc: func() time.Time { return deps.Now().UTC() }})

	tokens := security.NewTokenManager(cfg.JWT)
	userModule := buildUserModule(db, validator.New())

	metrics := observability.NewHTTPMetrics(prometheus.DefaultRegisterer)

	router := httpserver.New(cfg, func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}, tokens, metrics, log)

	apiV1 := router.Group("/api/v1")

	userModule.RegisterRoutes(apiV1, httpserver.Auth(tokens))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Application{
		server:  server,
		closeDB: closeDB,
	}, nil
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
