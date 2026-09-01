package user

import (
	"github.com/labstack/echo/v4"
	httpadapter "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/adapter/inbound/http"
	postgresadapter "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/adapter/outbound/postgres"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application"
	"gopkg.in/go-playground/validator.v9"
	"gorm.io/gorm"
	"time"
)

type Module struct{ handler *httpadapter.Handler }

func NewModule(db *gorm.DB, validate *validator.Validate) *Module {
	repo := postgresadapter.NewRepository(db)
	return &Module{handler: httpadapter.NewHandler(application.NewCreateUser(repo, time.Now), application.NewGetUser(repo), validate)}
}
func (m *Module) RegisterRoutes(group *echo.Group, auth echo.MiddlewareFunc) {
	m.handler.RegisterRoutes(group, auth)
}
