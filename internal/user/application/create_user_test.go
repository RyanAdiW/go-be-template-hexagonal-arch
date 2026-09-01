package application

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type fakeUsers struct{ users map[uuid.UUID]domain.User }

func (f *fakeUsers) Create(_ context.Context, user domain.User) error {
	for _, existing := range f.users {
		if existing.Email == user.Email {
			return domain.ErrEmailAlreadyExists
		}
	}
	f.users[user.ID] = user
	return nil
}
func (f *fakeUsers) GetByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}
func TestCreateUser(t *testing.T) {
	repo := &fakeUsers{users: map[uuid.UUID]domain.User{}}
	usecase := NewCreateUser(repo, func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
	user, err := usecase.Execute(context.Background(), CreateUserCommand{Email: "Ada@Example.com", Name: "Ada"})
	require.NoError(t, err)
	require.Equal(t, "ada@example.com", user.Email)
	_, err = usecase.Execute(context.Background(), CreateUserCommand{Email: "ada@example.com", Name: "Other"})
	require.True(t, errors.Is(err, domain.ErrEmailAlreadyExists))
}
