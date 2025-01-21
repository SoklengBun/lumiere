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
