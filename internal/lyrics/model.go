package lyrics

import (
	"lumiere/internal/artist"
	"lumiere/internal/models"
	"time"

	"gorm.io/gorm"
)

// Lyrics represents a song entry. It holds one primary title plus alt titles,
// multiple artists, multiple content versions (romaji/japanese/english/etc.),
// and related cover performances.
type Lyrics struct {
	ID        string         `json:"id" gorm:"primaryKey;size:64"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	// Primary display title.
	Title string `json:"title" gorm:"type:text;index"`

	AltTitles []string `json:"altTitles" gorm:"serializer:json;type:jsonb;default:'[]'"`

	// Relations
	Artists  []artist.Artist `json:"artists" gorm:"many2many:lyrics_artists;constraint:OnDelete:CASCADE"`
	Contents []LyricContent  `json:"contents" gorm:"foreignKey:LyricsID;constraint:OnDelete:CASCADE"`
	Covers   []LyricCover    `json:"covers" gorm:"foreignKey:LyricsID;references:ID;constraint:OnDelete:CASCADE"`

	// CreatedByID: which user added this lyrics entry (stored, not expanded).
	CreatedByID uint `json:"createdById" gorm:"index"`
}

// LyricCover stores a cover performance of a song. It only needs a video ID and performers.
type LyricCover struct {
	models.BaseModel

	LyricsID string          `json:"lyricsId" gorm:"index;size:64"`
	CoverID  string          `json:"id" gorm:"column:cover_id;size:64;index"`
	Artists  []artist.Artist `json:"artists" gorm:"many2many:lyric_cover_artists;constraint:OnDelete:CASCADE"`
}

// LyricContent stores different lyric versions (e.g., japanese, romaji, english).
type LyricContent struct {
	models.BaseModel

	LyricsID string `json:"lyricsId" gorm:"index;size:64"`
	Kind     string `json:"kind" gorm:"size:64"` // e.g., "japanese", "romaji", "english"
	Content  string `json:"content" gorm:"type:text"`
}
