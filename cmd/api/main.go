package main

import (
	userService "lumiere/cmd/api/service/user"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Load .env
	err := godotenv.Load()
	if err != nil {
		e.Logger.Fatal("Error loading .env file")
	}

	appHost := os.Getenv("APP.HOST")
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! ==> ")
	})

	e.POST("/register", userService.RegisterUser)

	e.Logger.Fatal(e.Start(appHost))
}
