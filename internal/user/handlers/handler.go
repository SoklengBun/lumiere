package handlers

import (
	usersvc "lumiere/internal/user/service"
	util "lumiere/internal/util"
	"strings"

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
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}

	resp, err := h.svc.Register(c.Request().Context(), usersvc.RegisterRequest{
		Username: b.Username, Password: b.Password, Name: b.Name,
	})
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, map[string]interface{}{"token": resp.Token, "user": resp.User})
}

func (h *Handler) Login(c echo.Context) error {
	var b loginRequestBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}

	resp, err := h.svc.Login(c.Request().Context(), usersvc.LoginRequest{
		Username: b.Username, Password: b.Password,
	})
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, map[string]interface{}{"token": resp.Token, "user": resp.User})
}

func (h *Handler) QuickLogin(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return util.JSONError(c, util.CodeBadRequest, "missing token")
	}

	token := auth
	// support `Bearer <token>` format
	if parts := strings.SplitN(auth, " ", 2); len(parts) == 2 {
		if strings.ToLower(parts[0]) == "bearer" {
			token = parts[1]
		}
	}

	user, err := h.svc.QuickLogin(c.Request().Context(), token)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, map[string]interface{}{"user": user})
}
