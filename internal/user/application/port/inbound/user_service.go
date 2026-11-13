package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
)

type CreateUserCommand struct {
	Email string
	Name  string
}

type UserServiceInterface interface {
	CreateUser(context.Context, CreateUserCommand) (domain.User, error)
	GetUserByID(context.Context, uuid.UUID) (domain.User, error)
}
