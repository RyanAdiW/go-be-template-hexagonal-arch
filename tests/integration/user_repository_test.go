//go:build integration

package integration_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/database"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	postgresadapter "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/adapter/outbound/db/users"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
)

func TestUserRepositoryWithPostgreSQL(t *testing.T) {
	ctx := context.Background()
	container, err := postgrescontainer.Run(ctx, "postgres:16-alpine",
		postgrescontainer.WithDatabase("app"),
		postgrescontainer.WithUsername("app"),
		postgrescontainer.WithPassword("app"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	migrator, err := migrate.New("file://../../migrations", dsn)
	require.NoError(t, err)
	require.NoError(t, migrator.Up())
	t.Cleanup(func() { _, _ = migrator.Close() })

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	repository := postgresadapter.NewUsersDB(postgresadapter.Config{Client: db})
	user, err := domain.NewUser("ada@example.com", "Ada", time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, repository.Create(ctx, user))

	got, err := repository.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.Email, got.Email)
	_, err = repository.GetByID(ctx, user.ID)
	require.NoError(t, err)
	t.Run("nested transaction rolls back all writes", func(t *testing.T) {
		transactor := database.NewTransactor(db)
		first, err := domain.NewUser("first@example.com", "First", time.Now().UTC())
		require.NoError(t, err)
		second, err := domain.NewUser("second@example.com", "Second", time.Now().UTC())
		require.NoError(t, err)
		failure := errors.New("workflow failed")
		err = transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := repository.Create(txCtx, first); err != nil {
				return err
			}
			return transactor.WithinTransaction(txCtx, func(nestedCtx context.Context) error {
				if err := repository.Create(nestedCtx, second); err != nil {
					return err
				}
				_, err := repository.GetByID(nestedCtx, first.ID)
				if err != nil {
					return err
				}
				return failure
			})
		})
		require.ErrorIs(t, err, failure)
		for _, id := range []uuid.UUID{first.ID, second.ID} {
			_, err := repository.GetByID(ctx, id)
			require.ErrorIs(t, err, domain.ErrUserNotFound)
		}
	})
	t.Run("transaction commits writes", func(t *testing.T) {
		user, err := domain.NewUser("committed@example.com", "Committed", time.Now().UTC())
		require.NoError(t, err)
		err = database.NewTransactor(db).WithinTransaction(ctx, func(txCtx context.Context) error {
			return repository.Create(txCtx, user)
		})
		require.NoError(t, err)
		got, err := repository.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, user.Email, got.Email)
	})

}
