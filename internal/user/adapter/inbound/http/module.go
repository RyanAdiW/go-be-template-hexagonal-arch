package httpadapter

import (
	"github.com/labstack/echo/v4"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/inbound"
	"gopkg.in/go-playground/validator.v9"
)

type Dependencies struct {
	UserService inbound.UserServiceInterface
	Validator   *validator.Validate
}

type UserModule struct {
	service  inbound.UserServiceInterface
	validate *validator.Validate
}

func NewModule(deps Dependencies) *UserModule {
	return &UserModule{
		service:  deps.UserService,
		validate: deps.Validator,
	}
}

func (m *UserModule) RegisterRoutes(group *echo.Group, auth echo.MiddlewareFunc) {
	group.POST("/users", m.Create)
	group.GET("/users/:id", m.Get, auth)
}
