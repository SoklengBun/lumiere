package repository

import (
	"context"
	"lumiere/internal/artist"
	"strings"

	"gorm.io/gorm"
)

type gormRepo struct{ db *gorm.DB }

func NewGormRepo(db *gorm.DB) ArtistRepo { return &gormRepo{db: db} }

func (r *gormRepo) Create(ctx context.Context, a *artist.Artist) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *gormRepo) Update(ctx context.Context, a *artist.Artist) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id uint) (*artist.Artist, error) {
	var a artist.Artist
	if err := r.db.WithContext(ctx).Preload("CV").First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormRepo) FindByIDs(ctx context.Context, ids []uint) ([]artist.Artist, error) {
	var list []artist.Artist
	if len(ids) == 0 {
		return list, nil
	}
	if err := r.db.WithContext(ctx).Preload("CV").Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByName searches artists by name. When searching for B, artists where B is a CV also appear.
// (Searching A does NOT surface B — that asymmetry is handled here by only traversing cv → artist direction.)
func (r *gormRepo) FindByName(ctx context.Context, q string) ([]artist.Artist, error) {
	var list []artist.Artist
	if q == "" {
		return list, nil
	}
	pattern := "%" + strings.ToLower(q) + "%"
	// Find direct matches OR artists whose CV matches the query.
	if err := r.db.WithContext(ctx).Preload("CV").
		Joins("LEFT JOIN artists cv_a ON cv_a.id = artists.cv_id").
		Where(
			"LOWER(artists.name) LIKE ? OR LOWER(artists.alt_name) LIKE ? OR LOWER(cv_a.name) LIKE ? OR LOWER(cv_a.alt_name) LIKE ?",
			pattern, pattern, pattern, pattern,
		).
		Distinct("artists.*").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) List(ctx context.Context) ([]artist.Artist, error) {
	var list []artist.Artist
	if err := r.db.WithContext(ctx).Preload("CV").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
