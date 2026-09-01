package httpadapter

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/security"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"time"
)

type Handler struct {
	create   *application.CreateUser
	get      *application.GetUser
	validate *validator.Validate
}

func NewHandler(create *application.CreateUser, get *application.GetUser, validate *validator.Validate) *Handler {
	return &Handler{create, get, validate}
}

type createUserRequest struct {
	Email string `json:"email" validate:"required,email,max=254"`
	Name  string `json:"name" validate:"required,min=1,max=100"`
}
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func present(user domain.User) userResponse {
	return userResponse{user.ID, user.Email, user.Name, user.CreatedAt.UTC().Format(time.RFC3339), user.UpdatedAt.UTC().Format(time.RFC3339)}
}
func (h *Handler) Create(c echo.Context) error {
	var request createUserRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	if err := h.validate.Struct(request); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "validation failed")
	}
	user, err := h.create.Execute(c.Request().Context(), application.CreateUserCommand{Email: request.Email, Name: request.Name})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": present(user)})
}
func (h *Handler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	identity := c.Get("identity").(security.Identity)
	if identity.UserID != id {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	user, err := h.get.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": present(user)})
}
