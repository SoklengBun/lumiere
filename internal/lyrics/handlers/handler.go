package handlers

import (
	"strconv"
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

type listLyricsResponse struct {
	models.BaseModel
	VideoID     string                   `json:"videoId"`
	Title       string                   `json:"title"`
	AltTitles   []string                 `json:"altTitles"`
	Artists     []artistmodel.Artist     `json:"artists"`
	Covers      []lyricsmodel.LyricCover `json:"covers"`
	CreatedByID uint                     `json:"createdById"`
}

type listResponse struct {
	Page   int                  `json:"page"`
	Offset int                  `json:"offset"`
	Total  int64                `json:"total"`
	Items  []listLyricsResponse `json:"items"`
}

type coverBody struct {
	ID        string `json:"id"`
	ArtistIDs []uint `json:"artistIds"`
}

type addBody struct {
	VideoID   string        `json:"videoId"`
	Title     string        `json:"title"`
	AltTitles []string      `json:"altTitles"`
	ArtistIDs []uint        `json:"artistIds"`
	Covers    []coverBody   `json:"covers"`
	Contents  []contentBody `json:"contents"`
}

type editBody struct {
	VideoID   *string        `json:"videoId"`
	Title     *string        `json:"title"`
	AltTitles *[]string      `json:"altTitles"`
	ArtistIDs *[]uint        `json:"artistIds"`
	Covers    *[]coverBody   `json:"covers"`
	Contents  *[]contentBody `json:"contents"`
}

func (h *Handler) Get(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return util.JSONError(c, util.CodeBadRequest, "invalid id")
	}

	l, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return util.JSONError(c, util.CodeNotFound, err.Error())
	}
	return util.JSONSuccess(c, l)
}

func (h *Handler) List(c echo.Context) error {
	page := 1
	if rawPage := strings.TrimSpace(c.QueryParam("page")); rawPage != "" {
		parsedPage, err := strconv.Atoi(rawPage)
		if err != nil || parsedPage < 1 {
			return util.JSONError(c, util.CodeBadRequest, "invalid page")
		}
		page = parsedPage
	}

	offset := 10
	if rawOffset := strings.TrimSpace(c.QueryParam("offset")); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 1 {
			return util.JSONError(c, util.CodeBadRequest, "invalid offset")
		}
		offset = parsedOffset
	}

	list, total, err := h.svc.List(c.Request().Context(), page, offset)
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	resp := make([]listLyricsResponse, 0, len(list))
	for _, lyric := range list {
		resp = append(resp, listLyricsResponse{
			BaseModel:   lyric.BaseModel,
			VideoID:     lyric.VideoID,
			Title:       lyric.Title,
			AltTitles:   lyric.AltTitles,
			Artists:     lyric.Artists,
			Covers:      lyric.Covers,
			CreatedByID: lyric.CreatedByID,
		})
	}

	return util.JSONSuccess(c, listResponse{
		Page:   page,
		Offset: offset,
		Total:  total,
		Items:  resp,
	})
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
	b.VideoID = strings.TrimSpace(b.VideoID)
	if b.VideoID == "" {
		return util.JSONError(c, util.CodeBadRequest, "videoId is required")
	}
	b.Title = strings.TrimSpace(b.Title)
	if b.Title == "" {
		return util.JSONError(c, util.CodeBadRequest, "title is required")
	}

	// build AltTitles
	var altTitles []string
	for _, t := range b.AltTitles {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		altTitles = append(altTitles, t)
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

	covers, err := h.resolveCovers(c, b.Covers, b.VideoID)
	if err != nil {
		return err
	}

	l := &lyricsmodel.Lyrics{
		VideoID:   b.VideoID,
		Title:     b.Title,
		AltTitles: altTitles,
		Artists:   artists,
		Covers:    covers,
		Contents:  contents,
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
	id, err := parseUintParam(c, "id")
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

	if b.VideoID != nil {
		videoID := strings.TrimSpace(*b.VideoID)
		if videoID == "" {
			return util.JSONError(c, util.CodeBadRequest, "videoId cannot be empty")
		}
		existing.VideoID = videoID
	}

	if b.Title != nil {
		title := strings.TrimSpace(*b.Title)
		if title == "" {
			return util.JSONError(c, util.CodeBadRequest, "title cannot be empty")
		}
		existing.Title = title
	}

	if b.AltTitles != nil {
		// Replace alt titles only when explicitly provided.
		altTitles := make([]string, 0, len(*b.AltTitles))
		for _, t := range *b.AltTitles {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			altTitles = append(altTitles, t)
		}
		existing.AltTitles = altTitles
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

	if b.Covers != nil {
		covers := []coverBody{}
		covers = *b.Covers

		resolvedCovers, err := h.resolveCovers(c, covers, existing.VideoID)
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

func (h *Handler) resolveCovers(c echo.Context, covers []coverBody, selfID string) ([]lyricsmodel.LyricCover, error) {
	if len(covers) == 0 {
		return nil, nil
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

func parseUintParam(c echo.Context, key string) (uint, error) {
	id64, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}
