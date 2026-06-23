package artist

import "lumiere/internal/models"

// Artist represents a singer or performer. Shared entity for many-to-many relations.
type Artist struct {
	models.BaseModel

	Name    string `json:"name" gorm:"size:255"`
	AltName string `json:"altName" gorm:"size:255"`

	// CVID is the ID of the real voice actor behind this artist (e.g. a character voiced by a person).
	CVID *uint   `json:"cvId" gorm:"index"`
	CV   *Artist `json:"cv" gorm:"foreignKey:CVID"`
}
