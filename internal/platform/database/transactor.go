package database

import (
	"context"

	shared "github.com/ryanadi/go-be-template-hexagonal-arch/internal/shared/application/port/outbound"
	"gorm.io/gorm"
)

type transactionContextKey struct{}

type Transactor struct {
	db *gorm.DB
}

var _ shared.Transactor = (*Transactor)(nil)

func NewTransactor(db *gorm.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := ctx.Value(transactionContextKey{}).(*gorm.DB); ok {
		return work(ctx)
	}

	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(context.WithValue(ctx, transactionContextKey{}, tx))
	})
}

func Connection(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(transactionContextKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}

	return db.WithContext(ctx)
}
