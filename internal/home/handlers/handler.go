package handlers

import (
	homesvc "lumiere/internal/home/service"
	lyricsmodel "lumiere/internal/lyrics"
	playlistmodel "lumiere/internal/playlist"
	util "lumiere/internal/util"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *homesvc.Service
}

func New(svc *homesvc.Service) *Handler {
	return &Handler{svc: svc}
}

type response struct {
	Songs     []playlistmodel.SongResponse     `json:"songs"`
	Playlists []playlistmodel.PlaylistResponse `json:"playlists"`
}

func (h *Handler) Get(c echo.Context) error {
	payload, err := h.svc.Get(c.Request().Context())
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	return util.JSONSuccess(c, response{
		Songs:     toHomeSongs(payload.Songs),
		Playlists: playlistmodel.ToPlaylistResponses(payload.Playlists, true),
	})
}

func toHomeSongs(list []lyricsmodel.Lyrics) []playlistmodel.SongResponse {
	out := make([]playlistmodel.SongResponse, 0, len(list))
	for _, song := range list {
		out = append(out, playlistmodel.ToSongResponse(song))
	}
	return out
}
