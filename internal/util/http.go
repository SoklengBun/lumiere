package util

import (
	"github.com/labstack/echo/v4"
)

// JSONError writes a standardized JSON error response. It always returns HTTP
// status 200 and uses negative codes to indicate failures as project rule.
func JSONError(c echo.Context, code int, msg string) error {
	return c.JSON(200, Response{Code: code, Message: MessageForCode(code, msg), Data: nil})
}

// JSONSuccess writes a standardized JSON success response with HTTP status 200.
func JSONSuccess(c echo.Context, data interface{}) error {
	return c.JSON(200, Response{Code: CodeSuccess, Message: MessageForCode(CodeSuccess, ""), Data: data})
}
