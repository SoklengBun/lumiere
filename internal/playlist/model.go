package playlist

import (
	lyrics "lumiere/internal/lyrics"
	"lumiere/internal/models"
)

// Playlist groups multiple lyrics entries into an ordered collection.
type Playlist struct {
	models.BaseModel

	Name        string `json:"name" gorm:"size:255;index"`
	Description string `json:"description" gorm:"type:text"`
	IsPublic    bool   `json:"isPublic" gorm:"index"`
	CreatedByID uint   `json:"createdById" gorm:"index"`

	Items []PlaylistItem `json:"items" gorm:"foreignKey:PlaylistID;constraint:OnDelete:CASCADE"`
}

// PlaylistItem points to one lyrics entry and keeps display order in a playlist.
type PlaylistItem struct {
	models.BaseModel

	PlaylistID     uint   `json:"playlistId" gorm:"index;uniqueIndex:idx_playlist_item_position"`
	LyricsID       uint   `json:"lyricsId" gorm:"index"`
	Position       uint   `json:"position" gorm:"uniqueIndex:idx_playlist_item_position"`
	Note           string `json:"note" gorm:"type:text"`
	DefaultCoverID string `json:"defaultCoverId" gorm:"column:default_cover_id;size:64;index"`

	Lyrics lyrics.Lyrics `json:"lyrics" gorm:"foreignKey:LyricsID;references:ID"`
}

type ItemOrder struct {
	ItemID   uint `json:"itemId"`
	Position uint `json:"position"`
}
