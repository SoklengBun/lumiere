package handlers

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.GET("/search", h.Search)
	g.GET("/:id", h.Get)
	g.GET("/list", h.List)
	g.GET("/mine", h.Mine)
	g.POST("/add", h.Add)
	g.PUT("/:id", h.Edit)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/items", h.AddItems)
	g.PUT("/:id/items/reorder", h.ReorderItems)
	g.DELETE("/:id/items/:itemId", h.DeleteItem)
}
