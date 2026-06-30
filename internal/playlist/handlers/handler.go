package handlers

import (
	"encoding/json"
	"errors"
	"lumiere/internal/models"
	playlistmodel "lumiere/internal/playlist"
	playlistsvc "lumiere/internal/playlist/service"
	usersvc "lumiere/internal/user/service"
	util "lumiere/internal/util"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc     *playlistsvc.Service
	userSvc *usersvc.Service
}

func New(svc *playlistsvc.Service, userSvc *usersvc.Service) *Handler {
	return &Handler{svc: svc, userSvc: userSvc}
}

type addBody struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsPublic    jsonBool `json:"isPublic"`
	LyricsIDs   []uint   `json:"lyricsIds"`
}

type jsonBool bool

func (b *jsonBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*b = false
		return nil
	}

	var boolValue bool
	if err := json.Unmarshal(data, &boolValue); err == nil {
		*b = jsonBool(boolValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err != nil {
		return err
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(stringValue))
	if err != nil {
		return err
	}

	*b = jsonBool(parsed)
	return nil
}

type reorderBody struct {
	ItemOrders []playlistmodel.ItemOrder `json:"itemOrders"`
}

type addItemsBody struct {
	LyricsIDs []uint `json:"lyricsIds"`
}

type updateItemBody struct {
	DefaultCoverID *string `json:"defaultCoverId"`
	Note           *string `json:"note"`
}

func (h *Handler) Get(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	p, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}

	if !p.IsPublic {
		user, err := h.authenticate(c)
		if err != nil {
			return util.JSONError(c, util.CodeUnauthorized, "")
		}
		if user.ID != p.CreatedByID {
			return util.JSONError(c, util.CodeFailed, "not allowed")
		}
	}

	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*p, false))
}

func (h *Handler) List(c echo.Context) error {
	list, err := h.svc.ListPublic(c.Request().Context())
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponses(list, true))
}

func (h *Handler) Search(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		return util.JSONError(c, util.CodeBadRequest, "missing query")
	}

	list, err := h.svc.Search(c.Request().Context(), q)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponses(list, true))
}

func (h *Handler) Mine(c echo.Context) error {
	user, err := h.authenticate(c)
	if err != nil {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	list, err := h.svc.ListByUser(c.Request().Context(), user.ID)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponses(list, false))
}

func (h *Handler) Add(c echo.Context) error {
	user, err := h.authenticate(c)
	if err != nil {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	var b addBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}
	if strings.TrimSpace(b.Name) == "" {
		return util.JSONError(c, util.CodeBadRequest, "name is required")
	}

	items, err := playlistItemsFromLyricsIDs(b.LyricsIDs)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}

	p := &playlistmodel.Playlist{
		Name:        b.Name,
		Description: b.Description,
		IsPublic:    bool(b.IsPublic),
		CreatedByID: user.ID,
		Items:       items,
	}

	created, err := h.svc.Create(c.Request().Context(), p)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*created, false))
}

func (h *Handler) Edit(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	user, p, err := h.authAndCheckOwnership(c, id)
	if err != nil {
		_ = user
		return err
	}

	var b addBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}
	if strings.TrimSpace(b.Name) == "" {
		return util.JSONError(c, util.CodeBadRequest, "name is required")
	}

	items, err := playlistItemsFromLyricsIDs(b.LyricsIDs)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}

	p.Name = b.Name
	p.Description = b.Description
	p.IsPublic = bool(b.IsPublic)
	p.Items = items

	updated, err := h.svc.Update(c.Request().Context(), p)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*updated, false))
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	_, _, authErr := h.authAndCheckOwnership(c, id)
	if authErr != nil {
		return authErr
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, map[string]bool{"deleted": true})
}

func (h *Handler) AddItems(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	_, _, authErr := h.authAndCheckOwnership(c, id)
	if authErr != nil {
		return authErr
	}

	var b addItemsBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}
	items, err := playlistItemsFromLyricsIDs(b.LyricsIDs)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	if len(items) == 0 {
		return util.JSONError(c, util.CodeBadRequest, "items is required")
	}

	if err := h.svc.AddItems(c.Request().Context(), id, items); err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}

	updated, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*updated, false))
}

func (h *Handler) UpdateItem(c echo.Context) error {
	itemID, err := parseUintParam(c, "itemId")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid item id")
	}

	user, err := h.authenticate(c)
	if err != nil {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	item, err := h.svc.GetItem(c.Request().Context(), itemID)
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}

	p, err := h.svc.Get(c.Request().Context(), item.PlaylistID)
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}
	if p.CreatedByID != user.ID {
		return util.JSONError(c, util.CodeFailed, "not allowed")
	}

	var b updateItemBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}
	if b.DefaultCoverID == nil && b.Note == nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}

	if err := h.svc.UpdateItem(c.Request().Context(), itemID, b.DefaultCoverID, b.Note); err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}

	updated, err := h.svc.Get(c.Request().Context(), item.PlaylistID)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*updated, false))
}

func (h *Handler) ReorderItems(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	_, _, authErr := h.authAndCheckOwnership(c, id)
	if authErr != nil {
		return authErr
	}

	var b reorderBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}

	if err := h.svc.ReorderItems(c.Request().Context(), id, b.ItemOrders); err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}

	updated, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*updated, false))
}

func (h *Handler) DeleteItem(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}
	itemID, err := parseUintParam(c, "itemId")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid item id")
	}

	_, _, authErr := h.authAndCheckOwnership(c, id)
	if authErr != nil {
		return authErr
	}

	if err := h.svc.DeleteItem(c.Request().Context(), id, itemID); err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	updated, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, playlistmodel.ToPlaylistResponse(*updated, false))
}

func (h *Handler) authenticate(c echo.Context) (*models.PublicUser, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return nil, echo.ErrUnauthorized
	}

	token := auth
	if parts := strings.SplitN(auth, " ", 2); len(parts) == 2 {
		if strings.ToLower(parts[0]) == "bearer" {
			token = parts[1]
		}
	}

	return h.userSvc.QuickLogin(c.Request().Context(), token)
}

func (h *Handler) authAndCheckOwnership(c echo.Context, id uint) (*models.PublicUser, *playlistmodel.Playlist, error) {
	user, err := h.authenticate(c)
	if err != nil {
		return nil, nil, util.JSONError(c, util.CodeUnauthorized, "")
	}

	p, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return user, nil, util.JSONError(c, util.CodeNotFound, err.Error())
	}

	if p.CreatedByID != user.ID {
		return user, p, util.JSONError(c, util.CodeFailed, "not allowed")
	}
	return user, p, nil
}

func parseUintParam(c echo.Context, key string) (uint, error) {
	idStr := c.Param(key)
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}

func playlistItemsFromLyricsIDs(lyricsIDs []uint) ([]playlistmodel.PlaylistItem, error) {
	out := make([]playlistmodel.PlaylistItem, 0, len(lyricsIDs))
	for i, lyricsID := range lyricsIDs {
		if lyricsID == 0 {
			return nil, errors.New("invalid lyrics id")
		}

		out = append(out, playlistmodel.PlaylistItem{
			LyricsID: lyricsID,
			Position: uint(i + 1),
		})
	}

	return out, nil
}
