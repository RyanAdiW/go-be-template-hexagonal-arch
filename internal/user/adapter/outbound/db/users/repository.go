package users

import (
	"context"
	"errors"

	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/database"
	repo "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/outbound/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"gorm.io/gorm"
)

type usersDB struct{ db *gorm.DB }
type Config struct{ Client *gorm.DB }

var _ repo.UserRepository = (*usersDB)(nil)

func NewUsersDB(config Config) *usersDB { return &usersDB{db: config.Client} }
func (r *usersDB) Create(ctx context.Context, user domain.User) error {
	err := database.Connection(ctx, r.db).Create(toModel(user)).Error
	if isUniqueViolation(err) {
		return domain.ErrEmailAlreadyExists
	}
	return err
}

func (r *usersDB) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var model userModel
	err := database.Connection(ctx, r.db).First(&model, "id = ?", id).Error
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
