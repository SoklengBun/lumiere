package artist

import "lumiere/internal/models"

// Artist represents a singer or performer. Shared entity for many-to-many relations.
type Artist struct {
	models.BaseModel

	Name           string `json:"name" gorm:"size:255;index"`
	NormalizedName string `json:"normalizedName" gorm:"size:255;index"`
}
