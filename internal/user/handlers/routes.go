package handlers

import "github.com/labstack/echo/v4"

// UserRoutes registers user routes on the provided group.
func UserRoutes(g *echo.Group, h *Handler) {
	// POST /register -> create/register user
	g.POST("/register", h.Register)
	// POST /login -> login
	g.POST("/login", h.Login)
}
