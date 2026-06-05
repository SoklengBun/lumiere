package handlers

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/:id", h.Get)
	g.GET("/list", h.List)
	g.GET("/mine", h.Mine)
	g.PUT("/:id", h.Edit)
	g.POST("/add", h.Add)
}
