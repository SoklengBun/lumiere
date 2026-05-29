package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRoleType uint

const (
	RoleUser       UserRoleType = 2
	RoleAdmin      UserRoleType = 1
	RoleSuperAdmin UserRoleType = 0
)

type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

type User struct {
	BaseModel
	Username string       `json:"username" gorm:"unique"`
	Password string       `json:"password"`
	Name     string       `json:"name"`
	Role     UserRoleType `json:"role" gorm:"default:2"`
}

// PublicUser is a sanitized representation safe for public responses.
type PublicUser struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Role      UserRoleType `json:"role"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// Public returns a PublicUser with sensitive fields removed.
func (u *User) Public() PublicUser {
	return PublicUser{
		ID:        u.ID,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
