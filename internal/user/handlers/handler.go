package handlers

import (
	"net/http"

	usersvc "lumiere/internal/user/service"
	util "lumiere/internal/util"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *usersvc.Service
}

func New(svc *usersvc.Service) *Handler { return &Handler{svc: svc} }

type registerRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(c echo.Context) error {
	var b registerRequestBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, http.StatusBadRequest, "missing params")
	}

	resp, err := h.svc.Register(c.Request().Context(), usersvc.RegisterRequest{
		Username: b.Username, Password: b.Password, Name: b.Name,
	})
	if err != nil {
		return util.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return util.JSONOK(c, map[string]interface{}{"token": resp.Token, "user": resp.User})
}

func (h *Handler) Login(c echo.Context) error {
	var b loginRequestBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, http.StatusBadRequest, "missing params")
	}

	resp, err := h.svc.Login(c.Request().Context(), usersvc.LoginRequest{
		Username: b.Username, Password: b.Password,
	})
	if err != nil {
		return util.JSONError(c, http.StatusBadRequest, err.Error())
	}
	return util.JSONOK(c, map[string]interface{}{"token": resp.Token, "user": resp.User})
}
