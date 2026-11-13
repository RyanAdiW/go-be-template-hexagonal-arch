package application

import (
	"time"

	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/inbound"
	repo "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/outbound/db"
)

type UserService struct {
	users repo.UserRepository
	now   func() time.Time
}

type Config struct {
	UserRepository repo.UserRepository
	Now            func() time.Time
}

var _ inbound.UserServiceInterface = (*UserService)(nil)

func NewUserService(config Config) *UserService {
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &UserService{
		users: config.UserRepository,
		now:   now,
	}
}
