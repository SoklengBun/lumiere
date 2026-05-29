package repository

import (
	"context"
	"lumiere/internal/artist"

	"gorm.io/gorm"
	"strings"
)

type gormRepo struct{ db *gorm.DB }

func NewGormRepo(db *gorm.DB) ArtistRepo { return &gormRepo{db: db} }

func (r *gormRepo) Create(ctx context.Context, a *artist.Artist) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id uint) (*artist.Artist, error) {
	var a artist.Artist
	if err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormRepo) FindByIDs(ctx context.Context, ids []uint) ([]artist.Artist, error) {
	var list []artist.Artist
	if len(ids) == 0 {
		return list, nil
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) FindByName(ctx context.Context, q string) ([]artist.Artist, error) {
	var list []artist.Artist
	if q == "" {
		return list, nil
	}
	// case-insensitive search by name
	pattern := "%" + strings.ToLower(q) + "%"
	if err := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? OR LOWER(normalized_name) LIKE ?", pattern, pattern).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) List(ctx context.Context) ([]artist.Artist, error) {
	var list []artist.Artist
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
