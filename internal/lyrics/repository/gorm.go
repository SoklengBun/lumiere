package repository

import (
	"context"
	artistmodel "lumiere/internal/artist"
	"lumiere/internal/lyrics"
	"strings"

	"gorm.io/gorm"
)

type gormRepo struct{ db *gorm.DB }

func NewGormRepo(db *gorm.DB) LyricsRepo { return &gormRepo{db: db} }

func preloadLyrics(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Artists").
		Preload("Artists.CV").
		Preload("Contents").
		Preload("Covers").
		Preload("Covers.Artists").
		Preload("Covers.Artists.CV")
}

func (r *gormRepo) Create(ctx context.Context, l *lyrics.Lyrics) error {
	// Ensure associations are saved as part of the create.
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(l).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id uint) (*lyrics.Lyrics, error) {
	var l lyrics.Lyrics
	if err := preloadLyrics(r.db.WithContext(ctx)).
		First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *gormRepo) GetByVideoID(ctx context.Context, videoID string) (*lyrics.Lyrics, error) {
	var l lyrics.Lyrics
	if err := preloadLyrics(r.db.WithContext(ctx)).
		First(&l, "video_id = ?", videoID).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *gormRepo) List(ctx context.Context, page int, offset int) ([]lyrics.Lyrics, int64, error) {
	var list []lyrics.Lyrics
	var total int64

	db := r.db.WithContext(ctx).Model(&lyrics.Lyrics{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("Artists").
		Preload("Artists.CV").
		Preload("Covers").
		Preload("Covers.Artists").
		Preload("Covers.Artists.CV").
		Order("id ASC").
		Limit(offset).
		Offset((page - 1) * offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *gormRepo) ListRandom(ctx context.Context, limit int) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if limit <= 0 {
		return list, nil
	}

	if err := r.db.WithContext(ctx).
		Preload("Artists").
		Preload("Artists.CV").
		Preload("Covers").
		Preload("Covers.Artists").
		Preload("Covers.Artists.CV").
		Order("RANDOM()").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (r *gormRepo) ListByUser(ctx context.Context, userID uint) ([]lyrics.Lyrics, error) {
	var list []lyrics.Lyrics
	if err := r.db.WithContext(ctx).
		Preload("Artists").
		Preload("Artists.CV").
		Preload("Contents").
		Preload("Covers").
		Preload("Covers.Artists").
		Preload("Covers.Artists.CV").
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
		Preload("Artists.CV").
		Preload("Covers").
		Preload("Covers.Artists").
		Preload("Covers.Artists.CV").
		Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (r *gormRepo) Update(ctx context.Context, l *lyrics.Lyrics) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		l.Artists = uniqueArtists(l.Artists)
		for i := range l.Covers {
			l.Covers[i].Artists = uniqueArtists(l.Covers[i].Artists)
		}

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
			Omit("Artists", "Artists.*", "Covers", "Covers.*", "Covers.Artists", "Covers.Artists.*").
			Save(l).Error; err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM lyrics_artists WHERE lyrics_id = ?", l.ID).Error; err != nil {
			return err
		}
		if len(l.Artists) > 0 {
			rows := make([]map[string]interface{}, 0, len(l.Artists))
			for _, artist := range l.Artists {
				rows = append(rows, map[string]interface{}{
					"lyrics_id": l.ID,
					"artist_id": artist.ID,
				})
			}
			if err := tx.Table("lyrics_artists").Create(&rows).Error; err != nil {
				return err
			}
		}

		var existingCovers []lyrics.LyricCover
		if err := tx.Where("lyrics_id = ?", l.ID).Find(&existingCovers).Error; err != nil {
			return err
		}

		existingByCoverID := make(map[string]lyrics.LyricCover, len(existingCovers))
		for _, cover := range existingCovers {
			existingByCoverID[cover.CoverID] = cover
		}

		desiredCoverIDs := make(map[string]struct{}, len(l.Covers))
		for i := range l.Covers {
			cover := &l.Covers[i]
			cover.LyricsID = l.ID
			desiredCoverIDs[cover.CoverID] = struct{}{}

			if existing, ok := existingByCoverID[cover.CoverID]; ok {
				cover.ID = existing.ID
			} else {
				artists := cover.Artists
				cover.Artists = nil
				if err := tx.Omit("Artists", "Artists.*").Create(cover).Error; err != nil {
					return err
				}
				cover.Artists = artists
			}

			if err := tx.Model(cover).Association("Artists").Replace(cover.Artists); err != nil {
				return err
			}
		}

		for _, existing := range existingCovers {
			if _, ok := desiredCoverIDs[existing.CoverID]; ok {
				continue
			}
			if err := tx.Model(&existing).Association("Artists").Clear(); err != nil {
				return err
			}
			if err := tx.Delete(&existing).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func uniqueArtists(artists []artistmodel.Artist) []artistmodel.Artist {
	seen := make(map[uint]struct{}, len(artists))
	out := make([]artistmodel.Artist, 0, len(artists))
	for _, artist := range artists {
		id := artist.ID
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, artist)
	}
	return out
}
