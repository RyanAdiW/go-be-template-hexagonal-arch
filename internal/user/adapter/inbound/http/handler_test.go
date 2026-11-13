package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/inbound"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/validator.v9"
)

type stubService struct {
	command inbound.CreateUserCommand
	calls   int
}

func (s *stubService) CreateUser(_ context.Context, command inbound.CreateUserCommand) (domain.User, error) {
	s.command = command
	s.calls++
	return domain.NewUser(command.Email, command.Name, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
}
func (s *stubService) GetUserByID(context.Context, uuid.UUID) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}

func TestCreateUserThroughInboundPort(t *testing.T) {
	for _, tc := range []struct {
		name, body    string
		status, calls int
	}{
		{"valid", `{"email":"ada@example.com","name":"Ada"}`, http.StatusCreated, 1},
		{"invalid JSON", `{`, http.StatusBadRequest, 0},
		{"invalid fields", `{"email":"invalid","name":""}`, http.StatusUnprocessableEntity, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &stubService{}
			module := NewModule(Dependencies{UserService: service, Validator: validator.New()})
			router := echo.New()
			module.RegisterRoutes(router.Group("/api/v1"), func(next echo.HandlerFunc) echo.HandlerFunc { return next })
			request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tc.body))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, tc.status, response.Code)
			require.Equal(t, tc.calls, service.calls)
			if tc.calls > 0 {
				require.Equal(t, inbound.CreateUserCommand{Email: "ada@example.com", Name: "Ada"}, service.command)
				require.Contains(t, response.Body.String(), `"data":`)
				require.Contains(t, response.Body.String(), `"email":"ada@example.com"`)
			}
		})
	}
}
