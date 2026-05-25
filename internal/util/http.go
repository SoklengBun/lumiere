package util

import (
	"github.com/labstack/echo/v4"
)

// JSONError writes a simple JSON error response.
func JSONError(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

// JSONOK writes a JSON success response with status 200.
func JSONOK(c echo.Context, data interface{}) error {
	return c.JSON(200, data)
}
