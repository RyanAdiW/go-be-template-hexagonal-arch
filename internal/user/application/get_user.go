package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
)

type GetUser struct{ users port.UserRepository }

func NewGetUser(users port.UserRepository) *GetUser { return &GetUser{users} }
func (u *GetUser) Execute(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return u.users.GetByID(ctx, id)
}
