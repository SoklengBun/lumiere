package handlers

import (
	"net/http"
	"strconv"
	"strings"

	artistmodel "lumiere/internal/artist"
	artistsvc "lumiere/internal/artist/service"
	lyricsmodel "lumiere/internal/lyrics"
	lyricssvc "lumiere/internal/lyrics/service"
	util "lumiere/internal/util"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc       *lyricssvc.Service
	artistSvc *artistsvc.Service
}

func New(svc *lyricssvc.Service, artistSvc *artistsvc.Service) *Handler {
	return &Handler{svc: svc, artistSvc: artistSvc}
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
		return util.JSONError(c, http.StatusBadRequest, "invalid id")
	}

	l, err := h.svc.Get(c.Request().Context(), uint(id64))
	if err != nil {
		return util.JSONError(c, http.StatusNotFound, err.Error())
	}
	return util.JSONOK(c, l)
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
			return util.JSONError(c, http.StatusInternalServerError, err.Error())
		}
		if len(found) != len(b.ArtistIDs) {
			return util.JSONError(c, http.StatusBadRequest, "one or more artist IDs are invalid")
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
	created, err := h.svc.Add(c.Request().Context(), l)
	if err != nil {
		return util.JSONError(c, http.StatusInternalServerError, err.Error())
	}
	return util.JSONOK(c, created)
}
