package handlers

import (
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
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	IsPublic      bool     `json:"isPublic"`
	ItemLyricsIDs []string `json:"itemLyricsIds"`
}

type reorderBody struct {
	ItemOrders []playlistmodel.ItemOrder `json:"itemOrders"`
}

type addItemsBody struct {
	LyricsIDs []string `json:"lyricsIds"`
}

type artistName struct {
	Name string `json:"name"`
}

type compactSong struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Artists []artistName `json:"artists"`
}

type compactPlaylistItem struct {
	ID       uint        `json:"id"`
	LyricsID string      `json:"lyricsId"`
	Position uint        `json:"position"`
	Song     compactSong `json:"song"`
}

type compactPlaylist struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	IsPublic    bool                  `json:"isPublic"`
	CreatedByID uint                  `json:"createdById"`
	Items       []compactPlaylistItem `json:"items"`
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

	return util.JSONSuccess(c, toCompactPlaylist(*p, false))
}

func (h *Handler) List(c echo.Context) error {
	list, err := h.svc.ListPublic(c.Request().Context())
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	out := make([]compactPlaylist, 0, len(list))
	for _, p := range list {
		out = append(out, toCompactPlaylist(p, true))
	}

	return util.JSONSuccess(c, out)
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

	out := make([]compactPlaylist, 0, len(list))
	for _, p := range list {
		out = append(out, toCompactPlaylist(p, true))
	}

	return util.JSONSuccess(c, out)
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
	return util.JSONSuccess(c, list)
}

func toCompactPlaylist(p playlistmodel.Playlist, homeMode bool) compactPlaylist {
	items := p.Items
	if homeMode && len(items) > 5 {
		items = items[:5]
	}

	outItems := make([]compactPlaylistItem, 0, len(items))
	for _, it := range items {
		artists := make([]artistName, 0, len(it.Lyrics.Artists))
		for _, a := range it.Lyrics.Artists {
			artists = append(artists, artistName{Name: a.Name})
		}

		name := strings.TrimSpace(it.Lyrics.Summary)
		if name == "" && len(it.Lyrics.Titles) > 0 {
			name = it.Lyrics.Titles[0].Title
		}

		outItems = append(outItems, compactPlaylistItem{
			ID:       it.ID,
			LyricsID: it.LyricsID,
			Position: it.Position,
			Song: compactSong{
				ID:      it.Lyrics.ID,
				Name:    name,
				Artists: artists,
			},
		})
	}

	return compactPlaylist{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsPublic:    p.IsPublic,
		CreatedByID: p.CreatedByID,
		Items:       outItems,
	}
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

	items := make([]playlistmodel.PlaylistItem, 0, len(b.ItemLyricsIDs))
	for i, lyricsID := range b.ItemLyricsIDs {
		lyricsID = strings.TrimSpace(lyricsID)
		if lyricsID == "" {
			return util.JSONError(c, util.CodeBadRequest, "invalid lyrics id")
		}
		items = append(items, playlistmodel.PlaylistItem{LyricsID: lyricsID, Position: uint(i + 1)})
	}

	p := &playlistmodel.Playlist{
		Name:        b.Name,
		Description: b.Description,
		IsPublic:    b.IsPublic,
		CreatedByID: user.ID,
		Items:       items,
	}

	created, err := h.svc.Create(c.Request().Context(), p)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, created)
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

	items := make([]playlistmodel.PlaylistItem, 0, len(b.ItemLyricsIDs))
	for i, lyricsID := range b.ItemLyricsIDs {
		lyricsID = strings.TrimSpace(lyricsID)
		if lyricsID == "" {
			return util.JSONError(c, util.CodeBadRequest, "invalid lyrics id")
		}
		items = append(items, playlistmodel.PlaylistItem{LyricsID: lyricsID, Position: uint(i + 1)})
	}

	p.Name = b.Name
	p.Description = b.Description
	p.IsPublic = b.IsPublic
	p.Items = items

	updated, err := h.svc.Update(c.Request().Context(), p)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}
	return util.JSONSuccess(c, updated)
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
	if len(b.LyricsIDs) == 0 {
		return util.JSONError(c, util.CodeBadRequest, "lyricsIds is required")
	}

	items := make([]playlistmodel.PlaylistItem, 0, len(b.LyricsIDs))
	for _, lyricsID := range b.LyricsIDs {
		lyricsID = strings.TrimSpace(lyricsID)
		if lyricsID == "" {
			return util.JSONError(c, util.CodeBadRequest, "invalid lyrics id")
		}
		items = append(items, playlistmodel.PlaylistItem{LyricsID: lyricsID})
	}

	if err := h.svc.AddItems(c.Request().Context(), id, items); err != nil {
		return util.JSONError(c, util.CodeBadRequest, err.Error())
	}

	updated, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, updated)
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
	return util.JSONSuccess(c, updated)
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
	return util.JSONSuccess(c, updated)
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
