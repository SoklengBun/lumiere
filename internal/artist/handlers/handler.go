package handlers

import (
	"net/http"
	"strconv"

	artistmodel "lumiere/internal/artist"
	artistsvc "lumiere/internal/artist/service"
	util "lumiere/internal/util"

	"github.com/labstack/echo/v4"
)

type Handler struct{ svc *artistsvc.Service }

func New(svc *artistsvc.Service) *Handler { return &Handler{svc: svc} }

type addBody struct {
	Name string `json:"name"`
}

func (h *Handler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return util.JSONError(c, http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetByID(c.Request().Context(), uint(id64))
	if err != nil {
		return util.JSONError(c, http.StatusNotFound, err.Error())
	}
	return util.JSONOK(c, a)
}

func (h *Handler) List(c echo.Context) error {
	list, err := h.svc.List(c.Request().Context())
	if err != nil {
		return util.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return util.JSONOK(c, list)
}

func (h *Handler) Add(c echo.Context) error {
	var b addBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, http.StatusBadRequest, "missing params")
	}
	a := &artistmodel.Artist{Name: b.Name}
	if err := h.svc.Create(c.Request().Context(), a); err != nil {
		return util.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return util.JSONOK(c, a)
}
