package playlist

import (
	artistmodel "lumiere/internal/artist"
	lyricsmodel "lumiere/internal/lyrics"
)

type SongResponse struct {
	ID        uint                     `json:"id"`
	VideoID   string                   `json:"videoId"`
	Title     string                   `json:"title"`
	AltTitles []string                 `json:"altTitles"`
	Artists   []artistmodel.Artist     `json:"artists"`
	Covers    []lyricsmodel.LyricCover `json:"covers"`
}

type ItemResponse struct {
	ID             uint         `json:"id"`
	LyricsID       uint         `json:"lyricsId"`
	DefaultCoverID string       `json:"defaultCoverId"`
	Position       uint         `json:"position"`
	Note           string       `json:"note"`
	Song           SongResponse `json:"song"`
}

type PlaylistResponse struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	IsPublic    bool           `json:"isPublic"`
	CreatedByID uint           `json:"createdById"`
	Items       []ItemResponse `json:"items"`
}

func ToPlaylistResponse(p Playlist, homeMode bool) PlaylistResponse {
	items := p.Items
	if homeMode && len(items) > 5 {
		items = items[:5]
	}

	outItems := make([]ItemResponse, 0, len(items))
	for _, item := range items {
		outItems = append(outItems, ItemResponse{
			ID:             item.ID,
			LyricsID:       item.LyricsID,
			DefaultCoverID: item.DefaultCoverID,
			Position:       item.Position,
			Note:           item.Note,
			Song:           ToSongResponse(item.Lyrics),
		})
	}

	return PlaylistResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsPublic:    p.IsPublic,
		CreatedByID: p.CreatedByID,
		Items:       outItems,
	}
}

func ToPlaylistResponses(list []Playlist, homeMode bool) []PlaylistResponse {
	out := make([]PlaylistResponse, 0, len(list))
	for _, p := range list {
		out = append(out, ToPlaylistResponse(p, homeMode))
	}
	return out
}

func ToSongResponse(song lyricsmodel.Lyrics) SongResponse {
	return SongResponse{
		ID:        song.ID,
		VideoID:   song.VideoID,
		Title:     song.Title,
		AltTitles: song.AltTitles,
		Artists:   song.Artists,
		Covers:    song.Covers,
	}
}
