package application

import (
	"context"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"time"
)

type CreateUser struct {
	users port.UserRepository
	now   func() time.Time
}

func NewCreateUser(users port.UserRepository, now func() time.Time) *CreateUser {
	return &CreateUser{users, now}
}
func (u *CreateUser) Execute(ctx context.Context, cmd CreateUserCommand) (domain.User, error) {
	user, err := domain.NewUser(cmd.Email, cmd.Name, u.now())
	if err != nil {
		return domain.User{}, err
	}
	if err = u.users.Create(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}
