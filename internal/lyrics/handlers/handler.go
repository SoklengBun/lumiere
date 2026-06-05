package handlers

import (
	"strconv"
	"strings"

	artistmodel "lumiere/internal/artist"
	artistsvc "lumiere/internal/artist/service"
	lyricsmodel "lumiere/internal/lyrics"
	lyricssvc "lumiere/internal/lyrics/service"
	usersvc "lumiere/internal/user/service"
	util "lumiere/internal/util"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc       *lyricssvc.Service
	artistSvc *artistsvc.Service
	userSvc   *usersvc.Service
}

func New(svc *lyricssvc.Service, artistSvc *artistsvc.Service, userSvc *usersvc.Service) *Handler {
	return &Handler{svc: svc, artistSvc: artistSvc, userSvc: userSvc}
}

type contentBody struct {
	Kind    string `json:"kind"`
	Lang    string `json:"lang"`
	Content string `json:"content"`
}

type referenceBody struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

type addBody struct {
	Titles     []string        `json:"titles"`
	ArtistIDs  []uint          `json:"artistIds"`
	Contents   []contentBody   `json:"contents"`
	References []referenceBody `json:"references"`
	Summary    string          `json:"summary"`
}

func (h *Handler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	l, err := h.svc.Get(c.Request().Context(), uint(id64))
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}
	return util.JSONSuccess(c, l)
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

	// build Titles
	var titles []lyricsmodel.LyricTitle
	for _, t := range b.Titles {
		titles = append(titles, lyricsmodel.LyricTitle{Title: t, Normalized: strings.ToLower(t)})
	}

	// resolve Artists by IDs
	var artists []artistmodel.Artist
	if len(b.ArtistIDs) > 0 {
		found, err := h.artistSvc.FindByIDs(c.Request().Context(), b.ArtistIDs)
		if err != nil {
			return util.JSONError(c, util.CodeInternal, err.Error())
		}
		if len(found) != len(b.ArtistIDs) {
			return util.JSONError(c, util.CodeBadRequest, "one or more artist IDs are invalid")
		}
		artists = found
	}

	// build Contents
	var contents []lyricsmodel.LyricContent
	for _, c := range b.Contents {
		contents = append(contents, lyricsmodel.LyricContent{Kind: c.Kind, Lang: c.Lang, Content: c.Content})
	}

	// build References
	var refs []lyricsmodel.LyricReference
	for _, r := range b.References {
		refs = append(refs, lyricsmodel.LyricReference{Link: r.Link, Name: r.Name})
	}

	l := &lyricsmodel.Lyrics{
		Summary:    b.Summary,
		Titles:     titles,
		Artists:    artists,
		Contents:   contents,
		References: refs,
	}
	// if an Authorization header is provided, resolve the user and attach CreatedByID
	auth := c.Request().Header.Get("Authorization")
	if auth != "" {
		token := auth
		if parts := strings.SplitN(auth, " ", 2); len(parts) == 2 {
			if strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}
		if user, err := h.userSvc.QuickLogin(c.Request().Context(), token); err == nil {
			l.CreatedByID = user.ID
		}
	}

	created, err := h.svc.Add(c.Request().Context(), l)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, created)
}

func (h *Handler) Mine(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	token := auth
	if parts := strings.SplitN(auth, " ", 2); len(parts) == 2 {
		if strings.ToLower(parts[0]) == "bearer" {
			token = parts[1]
		}
	}

	user, err := h.userSvc.QuickLogin(c.Request().Context(), token)
	if err != nil {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	list, err := h.svc.ListByUser(c.Request().Context(), user.ID)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	return util.JSONSuccess(c, list)
}

func (h *Handler) Edit(c echo.Context) error {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	token := auth
	if parts := strings.SplitN(auth, " ", 2); len(parts) == 2 {
		if strings.ToLower(parts[0]) == "bearer" {
			token = parts[1]
		}
	}

	user, err := h.userSvc.QuickLogin(c.Request().Context(), token)
	if err != nil {
		return util.JSONError(c, util.CodeUnauthorized, "")
	}

	existing, err := h.svc.Get(c.Request().Context(), uint(id64))
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}

	if existing.CreatedByID != user.ID {
		return util.JSONError(c, util.CodeFailed, "not allowed")
	}

	var b addBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}

	// build Titles
	var titles []lyricsmodel.LyricTitle
	for _, t := range b.Titles {
		titles = append(titles, lyricsmodel.LyricTitle{Title: t, Normalized: strings.ToLower(t)})
	}

	// resolve Artists by IDs
	var artists []artistmodel.Artist
	if len(b.ArtistIDs) > 0 {
		found, err := h.artistSvc.FindByIDs(c.Request().Context(), b.ArtistIDs)
		if err != nil {
			return util.JSONError(c, util.CodeInternal, err.Error())
		}
		if len(found) != len(b.ArtistIDs) {
			return util.JSONError(c, util.CodeBadRequest, "one or more artist IDs are invalid")
		}
		artists = found
	}

	// build Contents
	var contents []lyricsmodel.LyricContent
	for _, cbody := range b.Contents {
		contents = append(contents, lyricsmodel.LyricContent{Kind: cbody.Kind, Lang: cbody.Lang, Content: cbody.Content})
	}

	// build References
	var refs []lyricsmodel.LyricReference
	for _, r := range b.References {
		refs = append(refs, lyricsmodel.LyricReference{Link: r.Link, Name: r.Name})
	}

	existing.Summary = b.Summary
	existing.Titles = titles
	existing.Artists = artists
	existing.Contents = contents
	existing.References = refs

	updated, err := h.svc.Update(c.Request().Context(), existing)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, updated)
}
