package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(ctx context.Context, cfg config.DatabaseConfig) (*gorm.DB, func() error, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, nil, err
	}
	return db, sqlDB.Close, nil
}
func DSN(cfg config.DatabaseConfig) string {
	return (&url.URL{Scheme: "postgres", User: url.UserPassword(cfg.User, cfg.Password), Host: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Path: cfg.Name, RawQuery: "sslmode=" + cfg.SSLMode}).String()
}
