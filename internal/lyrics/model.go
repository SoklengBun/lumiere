package lyrics

import (
	"lumiere/internal/artist"
	"lumiere/internal/models"
	"time"

	"gorm.io/gorm"
)

// Lyrics represents a song entry. It holds relations to multiple titles,
// multiple artists, multiple content versions (romaji/japanese/english/etc.),
// and related cover performances.
type Lyrics struct {
	ID        string         `json:"id" gorm:"primaryKey;size:64"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	// Short description or primary title. Use Titles slice for multiple titles.
	Summary string `json:"summary" gorm:"type:text"`

	// Relations
	Titles   []LyricTitle    `json:"titles" gorm:"foreignKey:LyricsID;constraint:OnDelete:CASCADE"`
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

// LyricTitle stores alternative titles for a lyrics entry (e.g., Japanese, romanized).
type LyricTitle struct {
	models.BaseModel

	LyricsID   string `json:"lyricsId" gorm:"index;size:64"`
	Title      string `json:"title" gorm:"type:text"`
	Normalized string `json:"normalized" gorm:"index"`
	Lang       string `json:"lang" gorm:"size:32"`
}

// LyricContent stores different lyric versions (e.g., japanese, romaji, english).
type LyricContent struct {
	models.BaseModel

	LyricsID string `json:"lyricsId" gorm:"index;size:64"`
	Kind     string `json:"kind" gorm:"size:64"` // e.g., "japanese", "romaji", "english"
	Content  string `json:"content" gorm:"type:text"`
}
