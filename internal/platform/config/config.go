package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App           AppConfig `env-prefix:"APP_"`
	Server        ServerConfig
	Database      DatabaseConfig `env-prefix:"DATABASE_"`
	JWT           JWTConfig      `env-prefix:"JWT_"`
	Observability ObservabilityConfig
}
type AppConfig struct {
	Environment string `env:"ENV" env-default:"development"`
	Name        string `env:"NAME" env-default:"go-hex-template"`
}
type ServerConfig struct {
	Host               string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port               int           `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout        time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout       time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"15s"`
	IdleTimeout        time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout    time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
	MaxBodyBytes       int64         `env:"HTTP_MAX_BODY_BYTES" env-default:"1048576"`
	CORSAllowedOrigins string        `env:"CORS_ALLOWED_ORIGINS" env-default:"http://localhost:3000"`
}
type DatabaseConfig struct {
	Host            string        `env:"HOST"`
	Port            int           `env:"PORT" env-default:"5432"`
	User            string        `env:"USER"`
	Password        string        `env:"PASSWORD"`
	Name            string        `env:"NAME"`
	SSLMode         string        `env:"SSL_MODE" env-default:"require"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdleTime time.Duration `env:"CONN_MAX_IDLE_TIME" env-default:"5m"`
}
type JWTConfig struct {
	Secret         string        `env:"SECRET"`
	Issuer         string        `env:"ISSUER" env-default:"go-hex-template"`
	Audience       string        `env:"AUDIENCE" env-default:"go-hex-template-client"`
	AccessTokenTTL time.Duration `env:"ACCESS_TOKEN_TTL" env-default:"15m"`
}
type ObservabilityConfig struct {
	OTELEnabled          bool   `env:"OTEL_ENABLED" env-default:"false"`
	OTELExporterEndpoint string `env:"OTEL_EXPORTER_ENDPOINT"`
	MetricsEnabled       bool   `env:"METRICS_ENABLED" env-default:"true"`
	LogLevel             string `env:"LOG_LEVEL" env-default:"info"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.Database.Host == "" || cfg.Database.User == "" || cfg.Database.Name == "" {
		return Config{}, fmt.Errorf("database host, user, and name are required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	if cfg.Observability.OTELEnabled && cfg.Observability.OTELExporterEndpoint == "" {
		return Config{}, fmt.Errorf("OTEL_EXPORTER_ENDPOINT is required when OTEL_ENABLED=true")
	}
	if cfg.Server.Port < 1 || cfg.Database.Port < 1 || cfg.Database.MaxOpenConns < 1 || cfg.Database.MaxIdleConns < 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return Config{}, fmt.Errorf("invalid server or database connection configuration")
	}
	return cfg, nil
}
