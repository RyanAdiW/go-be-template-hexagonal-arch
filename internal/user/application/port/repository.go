package port

import (
	"context"
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
)

type UserRepository interface {
	Create(context.Context, domain.User) error
	GetByID(context.Context, uuid.UUID) (domain.User, error)
}
