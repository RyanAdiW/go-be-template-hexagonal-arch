package application

import (
	"context"

	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/inbound"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
)

func (u *UserService) CreateUser(ctx context.Context, cmd inbound.CreateUserCommand) (domain.User, error) {
	user, err := domain.NewUser(cmd.Email, cmd.Name, u.now())
	if err != nil {
		return domain.User{}, err
	}

	if err = u.users.Create(ctx, user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}
