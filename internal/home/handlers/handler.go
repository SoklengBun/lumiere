package handlers

import (
	artistmodel "lumiere/internal/artist"
	homesvc "lumiere/internal/home/service"
	lyricsmodel "lumiere/internal/lyrics"
	playlistmodel "lumiere/internal/playlist"
	util "lumiere/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *homesvc.Service
}

func New(svc *homesvc.Service) *Handler {
	return &Handler{svc: svc}
}

type homeSong struct {
	ID        uint                     `json:"id"`
	VideoID   string                   `json:"videoId"`
	Title     string                   `json:"title"`
	AltTitles []string                 `json:"altTitles"`
	Artists   []artistmodel.Artist     `json:"artists"`
	Covers    []lyricsmodel.LyricCover `json:"covers"`
}

type compactSong struct {
	ID      uint                 `json:"id"`
	VideoID string               `json:"videoId"`
	Name    string               `json:"name"`
	Artists []artistmodel.Artist `json:"artists"`
}

type compactPlaylistItem struct {
	ID             uint        `json:"id"`
	LyricsID       uint        `json:"lyricsId"`
	DefaultCoverID string      `json:"defaultCoverId"`
	Position       uint        `json:"position"`
	Note           string      `json:"note"`
	Song           compactSong `json:"song"`
}

type compactPlaylist struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	IsPublic    bool                  `json:"isPublic"`
	CreatedByID uint                  `json:"createdById"`
	Items       []compactPlaylistItem `json:"items"`
}

type response struct {
	Songs     []homeSong        `json:"songs"`
	Playlists []compactPlaylist `json:"playlists"`
}

func (h *Handler) Get(c echo.Context) error {
	payload, err := h.svc.Get(c.Request().Context())
	if err != nil {
		return util.JSONError(c, util.CodeInternal, err.Error())
	}

	return util.JSONSuccess(c, response{
		Songs:     toHomeSongs(payload.Songs),
		Playlists: toCompactPlaylists(payload.Playlists),
	})
}

func toHomeSongs(list []lyricsmodel.Lyrics) []homeSong {
	out := make([]homeSong, 0, len(list))
	for _, song := range list {
		out = append(out, homeSong{
			ID:        song.ID,
			VideoID:   song.VideoID,
			Title:     song.Title,
			AltTitles: song.AltTitles,
			Artists:   song.Artists,
			Covers:    song.Covers,
		})
	}
	return out
}

func toCompactPlaylists(list []playlistmodel.Playlist) []compactPlaylist {
	out := make([]compactPlaylist, 0, len(list))
	for _, playlist := range list {
		items := make([]compactPlaylistItem, 0, len(playlist.Items))
		for _, item := range playlist.Items {
			items = append(items, compactPlaylistItem{
				ID:             item.ID,
				LyricsID:       item.LyricsID,
				DefaultCoverID: item.DefaultCoverID,
				Position:       item.Position,
				Note:           item.Note,
				Song:           toCompactSong(item.Lyrics),
			})
		}

		out = append(out, compactPlaylist{
			ID:          playlist.ID,
			Name:        playlist.Name,
			Description: playlist.Description,
			IsPublic:    playlist.IsPublic,
			CreatedByID: playlist.CreatedByID,
			Items:       items,
		})
	}
	return out
}

func toCompactSong(song lyricsmodel.Lyrics) compactSong {
	name := strings.TrimSpace(song.Title)
	if name == "" && len(song.AltTitles) > 0 {
		name = song.AltTitles[0]
	}

	return compactSong{
		ID:      song.ID,
		VideoID: song.VideoID,
		Name:    name,
		Artists: song.Artists,
	}
}
