package repository

import (
	"context"
	"lumiere/internal/lyrics"

	"gorm.io/gorm"
)

type gormRepo struct{ db *gorm.DB }

func NewGormRepo(db *gorm.DB) LyricsRepo { return &gormRepo{db: db} }

func (r *gormRepo) Create(ctx context.Context, l *lyrics.Lyrics) error {
	// Ensure associations are saved as part of the create.
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(l).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id uint) (*lyrics.Lyrics, error) {
	var l lyrics.Lyrics
	if err := r.db.WithContext(ctx).
		Preload("Titles").
		Preload("Artists").
		Preload("Contents").
		Preload("References").
		First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *gormRepo) List(ctx context.Context) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if err := r.db.WithContext(ctx).
		Preload("Titles").
		Preload("Artists").
		Preload("Contents").
		Preload("References").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) ListByUser(ctx context.Context, userID uint) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if err := r.db.WithContext(ctx).
		Preload("Titles").
		Preload("Artists").
		Preload("Contents").
		Preload("References").
		Where("created_by_id = ?", userID).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) Update(ctx context.Context, l *lyrics.Lyrics) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(l).Error
}
