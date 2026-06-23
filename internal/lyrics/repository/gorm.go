package repository

import (
	"context"
	"lumiere/internal/lyrics"
	"strings"

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
		Preload("Artists").
		Preload("Contents").
		Preload("Covers").
		Preload("Covers.Artists").
		First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *gormRepo) List(ctx context.Context) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if err := r.db.WithContext(ctx).
		Preload("Artists").
		Preload("Contents").
		Preload("Covers").
		Preload("Covers.Artists").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) ListByUser(ctx context.Context, userID uint) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if err := r.db.WithContext(ctx).
		Preload("Artists").
		Preload("Contents").
		Preload("Covers").
		Preload("Covers.Artists").
		Where("created_by_id = ?", userID).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) Search(ctx context.Context, q string) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if q == "" {
		return list, nil
	}

	pattern := "%" + strings.ToLower(q) + "%"
	if err := r.db.WithContext(ctx).
		Model(&lyrics.Lyrics{}).
		Distinct("lyrics.*").
		Joins("LEFT JOIN lyrics_artists ON lyrics_artists.lyrics_id = lyrics.id").
		Joins("LEFT JOIN artists ON artists.id = lyrics_artists.artist_id").
		Where(
			`LOWER(lyrics.title) LIKE ?
				OR LOWER(lyrics.video_id) LIKE ?
				OR EXISTS (
					SELECT 1
					FROM jsonb_array_elements_text(COALESCE(lyrics.alt_titles, '[]'::jsonb)) AS alt_title
					WHERE LOWER(alt_title) LIKE ?
				)
				OR LOWER(artists.name) LIKE ?`,
			pattern,
			pattern,
			pattern,
			pattern,
		).
		Order("lyrics.updated_at DESC").
		Limit(20).
		Preload("Artists").
		Preload("Contents").
		Preload("Covers").
		Preload("Covers.Artists").
		Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (r *gormRepo) Update(ctx context.Context, l *lyrics.Lyrics) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(l).Error; err != nil {
			return err
		}

		// Force exact replacement for many2many artists; Save can otherwise retain stale links.
		if err := tx.Model(l).Association("Artists").Replace(l.Artists); err != nil {
			return err
		}

		return nil
	})
}
