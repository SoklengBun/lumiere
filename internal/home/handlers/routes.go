package handlers

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("", h.Get)
	g.GET("/", h.Get)
}
