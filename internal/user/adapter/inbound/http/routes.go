package httpadapter

import "github.com/labstack/echo/v4"

func (h *Handler) RegisterRoutes(group *echo.Group, auth echo.MiddlewareFunc) {
	group.POST("/users", h.Create)
	group.GET("/users/:id", h.Get, auth)
}
