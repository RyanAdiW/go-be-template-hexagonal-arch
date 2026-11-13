package httpadapter

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/security"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application/port/inbound"
)

func (h *UserModule) Create(c echo.Context) error {
	var request createUserRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}

	if err := h.validate.Struct(request); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "validation failed")
	}

	user, err := h.service.CreateUser(c.Request().Context(), inbound.CreateUserCommand{
		Email: request.Email,
		Name:  request.Name,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]any{"data": present(user)})
}

func (h *UserModule) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	identity := c.Get("identity").(security.Identity)
	if identity.UserID != id {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{"data": present(user)})
}
