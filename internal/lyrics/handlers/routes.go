package handlers

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/:id", h.Get)
	g.GET("/list", h.List)
	g.POST("/add", h.Add)
}
