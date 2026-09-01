package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }
func (r *Repository) Create(ctx context.Context, user domain.User) error {
	err := r.db.WithContext(ctx).Create(toModel(user)).Error
	if isUniqueViolation(err) {
		return domain.ErrEmailAlreadyExists
	}
	return err
}
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return toDomain(model), nil
}
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
