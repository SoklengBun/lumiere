package userService

import (
	"encoding/json"
	"lumiere/internal/database"
	"lumiere/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Code string          `json:"code"`
	Data json.RawMessage `json:"data"`
}

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func RegisterUser(c echo.Context) error {

	db, err := database.Connect()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Code: "-1", Message: "Failed to connect database"})
	}

	u := new(Request)
	if err := c.Bind(u); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Code: "-1", Message: "Missing params"})
	}

	user := models.User{}
	result := db.First(&user, "username = ?", u.Username)

	if result.Error == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Code: "-1", Message: "Username already exist"})
	}

	user.Name = u.Name
	user.Username = u.Username
	user.Password = u.Password

	db.Create(&user)
	userJson, err := json.Marshal(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Code: "-2", Message: "wrong json"})
	}

	return c.JSON(http.StatusOK, Response{Code: "1", Data: userJson})
}
