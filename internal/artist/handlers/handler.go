package handlers

import (
	"encoding/json"
	"strconv"

	artistmodel "lumiere/internal/artist"
	artistsvc "lumiere/internal/artist/service"
	util "lumiere/internal/util"

	"github.com/labstack/echo/v4"
)

type Handler struct{ svc *artistsvc.Service }

func New(svc *artistsvc.Service) *Handler { return &Handler{svc: svc} }

type addBody struct {
	Name    string `json:"name"`
	AltName string `json:"altName"`
}

type optionalString struct {
	Set   bool
	Value string
}

func (o *optionalString) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = ""
		return nil
	}
	return json.Unmarshal(data, &o.Value)
}

type updateBody struct {
	Name    string         `json:"name"`
	AltName optionalString `json:"altName"`
	CVID    *uint          `json:"cvId"`
}

func (h *Handler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}
	a, err := h.svc.GetByID(c.Request().Context(), uint(id64))
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}
	return util.JSONSuccess(c, a)
}

func (h *Handler) List(c echo.Context) error {
	list, err := h.svc.List(c.Request().Context())
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, list)
}

func (h *Handler) Add(c echo.Context) error {
	var b addBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}
	a := &artistmodel.Artist{Name: b.Name, AltName: b.AltName}
	if err := h.svc.Create(c.Request().Context(), a); err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, a)
}

func (h *Handler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}
	existing, err := h.svc.GetByID(c.Request().Context(), uint(id64))
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, "artist not found")
	}
	var b updateBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}
	if b.Name != "" {
		existing.Name = b.Name
	}
	if b.AltName.Set {
		existing.AltName = b.AltName.Value
	}
	existing.CVID = b.CVID
	if err := h.svc.Update(c.Request().Context(), existing); err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, existing)
}

func (h *Handler) Search(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return util.JSONError(c, util.CodeBadRequest, "missing query")
	}
	list, err := h.svc.FindByName(c.Request().Context(), q)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, list)
}
