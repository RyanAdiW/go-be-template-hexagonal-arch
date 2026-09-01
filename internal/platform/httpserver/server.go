package httpserver

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/observability"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/security"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"go.opentelemetry.io/otel"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func New(cfg config.Config, ready func(context.Context) error, token *security.TokenManager, metrics *observability.HTTPMetrics, log *slog.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = errorHandler(log)
	e.Use(middleware.RequestID(), middleware.Recover(), middleware.BodyLimit("1M"), middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: strings.Split(cfg.Server.CORSAllowedOrigins, ","), AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderXRequestID}}), request(log, metrics), headers)
	e.GET("/health/live", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
	e.GET("/health/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := ready(ctx); err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "service not ready")
		}
		return c.NoContent(http.StatusNoContent)
	})
	if cfg.Observability.MetricsEnabled {
		e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	}
	return e
}
func Auth(token *security.TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			parts := strings.Fields(c.Request().Header.Get(echo.HeaderAuthorization))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}
			identity, err := token.Verify(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
			}
			c.Set("identity", identity)
			return next(c)
		}
	}
}
func headers(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("X-Content-Type-Options", "nosniff")
		c.Response().Header().Set("X-Frame-Options", "DENY")
		c.Response().Header().Set("Referrer-Policy", "no-referrer")
		return next(c)
	}
}
func request(log *slog.Logger, metrics *observability.HTTPMetrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			ctx, span := otel.Tracer("http").Start(c.Request().Context(), c.Request().Method+" "+c.Path())
			defer span.End()
			c.SetRequest(c.Request().WithContext(ctx))
			err := next(c)
			metrics.Observe(c.Request().Method, c.Path(), c.Response().Status, time.Since(start))
			log.Info("request", "request_id", c.Response().Header().Get(echo.HeaderXRequestID), "method", c.Request().Method, "path", c.Path(), "status", c.Response().Status, "duration", time.Since(start))
			return err
		}
	}
}
func errorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "internal server error"
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			status = httpErr.Code
			switch status {
			case http.StatusUnauthorized:
				code, message = "UNAUTHENTICATED", "authentication required"
			case http.StatusForbidden:
				code, message = "FORBIDDEN", "forbidden"
			case http.StatusUnprocessableEntity:
				code, message = "VALIDATION_FAILED", "validation failed"
			default:
				code, message = "INVALID_REQUEST", "invalid request"
			}
		} else if errors.Is(err, domain.ErrUserNotFound) {
			status = http.StatusNotFound
			code = "USER_NOT_FOUND"
			message = "user not found"
		} else if errors.Is(err, domain.ErrEmailAlreadyExists) {
			status = http.StatusConflict
			code = "EMAIL_ALREADY_EXISTS"
			message = "email already exists"
		} else {
			log.Error("unhandled request error", "error", err)
		}
		_ = c.JSON(status, map[string]any{"error": map[string]any{"code": code, "message": message, "details": nil}, "request_id": c.Response().Header().Get(echo.HeaderXRequestID)})
	}
}
