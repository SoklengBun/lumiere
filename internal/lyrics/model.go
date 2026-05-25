package lyrics

import (
	"lumiere/internal/artist"
	"lumiere/internal/models"
)

// Lyrics represents a song entry. It holds relations to multiple titles,
// multiple artists, multiple content versions (romaji/japanese/english/etc.),
// and reference links (YouTube, channels).
type Lyrics struct {
	models.BaseModel

	// Short description or primary title. Use Titles slice for multiple titles.
	Summary string `json:"summary" gorm:"type:text"`

	// Relations
	Titles     []LyricTitle     `json:"titles" gorm:"foreignKey:LyricsID;constraint:OnDelete:CASCADE"`
	Artists    []artist.Artist  `json:"artists" gorm:"many2many:lyrics_artists;constraint:OnDelete:CASCADE"`
	Contents   []LyricContent   `json:"contents" gorm:"foreignKey:LyricsID;constraint:OnDelete:CASCADE"`
	References []LyricReference `json:"references" gorm:"foreignKey:LyricsID;constraint:OnDelete:CASCADE"`
}

// LyricTitle stores alternative titles for a lyrics entry (e.g., Japanese, romanized).
type LyricTitle struct {
	models.BaseModel

	LyricsID   uint   `json:"lyricsId" gorm:"index"`
	Title      string `json:"title" gorm:"type:text"`
	Normalized string `json:"normalized" gorm:"index"`
	Lang       string `json:"lang" gorm:"size:32"`
}

// LyricContent stores different lyric versions (romaji, Japanese, pinyin, English, etc.).
type LyricContent struct {
	models.BaseModel

	LyricsID uint   `json:"lyricsId" gorm:"index"`
	Kind     string `json:"kind" gorm:"size:64"` // e.g., "romaji", "japanese", "english"
	Lang     string `json:"lang" gorm:"size:32"`
	Content  string `json:"content" gorm:"type:text"`
}

// LyricReference stores external reference links for a lyrics entry (e.g., YouTube videos).
type LyricReference struct {
	models.BaseModel

	LyricsID uint   `json:"lyricsId" gorm:"index"`
	Link     string `json:"link" gorm:"type:text"`
	Name     string `json:"name" gorm:"size:255"`
}
