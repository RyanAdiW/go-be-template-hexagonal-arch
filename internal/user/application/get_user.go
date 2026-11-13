package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
)

func (u *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return u.users.GetByID(ctx, id)
}
