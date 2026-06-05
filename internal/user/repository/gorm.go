package repository

import (
	"context"
	"lumiere/internal/models"

	"gorm.io/gorm"
)

type gormRepo struct {
	db *gorm.DB
}

func NewGormRepo(db *gorm.DB) UserRepo { return &gormRepo{db: db} }

func (r *gormRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *gormRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
