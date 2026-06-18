package handlers

import (
	"strings"

	artistmodel "lumiere/internal/artist"
	artistsvc "lumiere/internal/artist/service"
	lyricsmodel "lumiere/internal/lyrics"
	lyricssvc "lumiere/internal/lyrics/service"
	"lumiere/internal/models"
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
	Content string `json:"content"`
}

type coverBody struct {
	ID        string `json:"id"`
	ArtistIDs []uint `json:"artistIds"`
}

type addBody struct {
	ID        string        `json:"id"`
	Titles    []string      `json:"titles"`
	ArtistIDs []uint        `json:"artistIds"`
	Covers    []coverBody   `json:"covers"`
	CoverIDs  []string      `json:"coverIds"`
	Contents  []contentBody `json:"contents"`
	Summary   string        `json:"summary"`
}

type editBody struct {
	Titles    *[]string      `json:"titles"`
	ArtistIDs *[]uint        `json:"artistIds"`
	Covers    *[]coverBody   `json:"covers"`
	CoverIDs  *[]string      `json:"coverIds"`
	Contents  *[]contentBody `json:"contents"`
	Summary   *string        `json:"summary"`
}

func (h *Handler) Get(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	l, err := h.svc.Get(c.Request().Context(), id)
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

func (h *Handler) Search(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		return util.JSONError(c, util.CodeBadRequest, "missing query")
	}

	list, err := h.svc.Search(c.Request().Context(), q)
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
	b.ID = strings.TrimSpace(b.ID)
	if b.ID == "" {
		return util.JSONError(c, util.CodeBadRequest, "id is required")
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
		contents = append(contents, lyricsmodel.LyricContent{Kind: c.Kind, Content: c.Content})
	}

	covers, err := h.resolveCovers(c, b.Covers, b.CoverIDs, b.ID)
	if err != nil {
		return err
	}

	l := &lyricsmodel.Lyrics{
		ID:       b.ID,
		Summary:  b.Summary,
		Titles:   titles,
		Artists:  artists,
		Covers:   covers,
		Contents: contents,
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
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
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

	existing, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}

	if existing.CreatedByID != user.ID && user.Role != models.RoleSuperAdmin {
		return util.JSONError(c, util.CodeFailed, "not allowed")
	}

	var b editBody
	if err := c.Bind(&b); err != nil {
		return util.JSONError(c, util.CodeBadRequest, "missing params")
	}

	if b.Summary != nil {
		existing.Summary = *b.Summary
	}

	if b.Titles != nil {
		// Replace titles only when explicitly provided.
		titles := make([]lyricsmodel.LyricTitle, 0, len(*b.Titles))
		for _, t := range *b.Titles {
			titles = append(titles, lyricsmodel.LyricTitle{Title: t, Normalized: strings.ToLower(t)})
		}
		existing.Titles = titles
	}

	if b.ArtistIDs != nil {
		// Replace artists only when explicitly provided.
		artists := make([]artistmodel.Artist, 0, len(*b.ArtistIDs))
		if len(*b.ArtistIDs) > 0 {
			found, err := h.artistSvc.FindByIDs(c.Request().Context(), *b.ArtistIDs)
			if err != nil {
				return util.JSONError(c, util.CodeInternal, err.Error())
			}
			if len(found) != len(*b.ArtistIDs) {
				return util.JSONError(c, util.CodeBadRequest, "one or more artist IDs are invalid")
			}
			artists = found
		}
		existing.Artists = artists
	}

	if b.Contents != nil {
		// Replace contents only when explicitly provided.
		contents := make([]lyricsmodel.LyricContent, 0, len(*b.Contents))
		for _, cbody := range *b.Contents {
			contents = append(contents, lyricsmodel.LyricContent{Kind: cbody.Kind, Content: cbody.Content})
		}
		existing.Contents = contents
	}

	if b.Covers != nil || b.CoverIDs != nil {
		covers := []coverBody{}
		coverIDs := []string{}
		if b.Covers != nil {
			covers = *b.Covers
		}
		if b.CoverIDs != nil {
			coverIDs = *b.CoverIDs
		}

		resolvedCovers, err := h.resolveCovers(c, covers, coverIDs, existing.ID)
		if err != nil {
			return err
		}
		existing.Covers = resolvedCovers
	}

	updated, err := h.svc.Update(c.Request().Context(), existing)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}
	return util.JSONSuccess(c, updated)
}

func (h *Handler) resolveCovers(c echo.Context, covers []coverBody, legacyCoverIDs []string, selfID string) ([]lyricsmodel.LyricCover, error) {
	if len(covers) == 0 && len(legacyCoverIDs) == 0 {
		return nil, nil
	}

	if len(covers) == 0 && len(legacyCoverIDs) > 0 {
		covers = make([]coverBody, 0, len(legacyCoverIDs))
		for _, coverID := range legacyCoverIDs {
			covers = append(covers, coverBody{ID: coverID})
		}
	}

	seen := make(map[string]struct{}, len(covers))
	out := make([]lyricsmodel.LyricCover, 0, len(covers))
	for _, coverBody := range covers {
		coverID := strings.TrimSpace(coverBody.ID)
		if coverID == "" {
			return nil, util.JSONError(c, util.CodeBadRequest, "invalid cover id")
		}
		if coverID == selfID {
			return nil, util.JSONError(c, util.CodeBadRequest, "cover id cannot reference itself")
		}
		if _, ok := seen[coverID]; ok {
			continue
		}
		seen[coverID] = struct{}{}

		artists := make([]artistmodel.Artist, 0, len(coverBody.ArtistIDs))
		if len(coverBody.ArtistIDs) > 0 {
			found, err := h.artistSvc.FindByIDs(c.Request().Context(), coverBody.ArtistIDs)
			if err != nil {
				return nil, util.JSONError(c, util.CodeInternal, err.Error())
			}
			if len(found) != len(coverBody.ArtistIDs) {
				return nil, util.JSONError(c, util.CodeBadRequest, "one or more cover artist IDs are invalid")
			}
			artists = found
		}

		out = append(out, lyricsmodel.LyricCover{
			CoverID: coverID,
			Artists: artists,
		})
	}

	return out, nil
}
