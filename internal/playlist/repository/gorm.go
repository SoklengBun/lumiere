package repository

import (
	"context"
	"lumiere/internal/playlist"
	"strings"

	"gorm.io/gorm"
)

type gormRepo struct{ db *gorm.DB }

func NewGormRepo(db *gorm.DB) PlaylistRepo { return &gormRepo{db: db} }

func (r *gormRepo) preload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Items", func(tx *gorm.DB) *gorm.DB { return tx.Order("position ASC") }).
		Preload("Items.Lyrics").
		Preload("Items.Lyrics.Artists").
		Preload("Items.Lyrics.Artists.CV").
		Preload("Items.Lyrics.Covers").
		Preload("Items.Lyrics.Covers.Artists").
		Preload("Items.Lyrics.Covers.Artists.CV")
}

func (r *gormRepo) Create(ctx context.Context, p *playlist.Playlist) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(p).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id uint) (*playlist.Playlist, error) {
	var p playlist.Playlist
	if err := r.preload(r.db.WithContext(ctx)).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *gormRepo) GetItemByID(ctx context.Context, itemID uint) (*playlist.PlaylistItem, error) {
	var item playlist.PlaylistItem
	if err := r.db.WithContext(ctx).
		Preload("Lyrics").
		Preload("Lyrics.Covers").
		First(&item, "id = ?", itemID).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormRepo) ListPublic(ctx context.Context) ([]playlist.Playlist, error) {
	var list []playlist.Playlist
	if err := r.preload(r.db.WithContext(ctx)).
		Where("is_public = ?", true).
		Order("RANDOM()").
		Limit(5).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) SearchByName(ctx context.Context, q string) ([]playlist.Playlist, error) {
	var list []playlist.Playlist
	if q == "" {
		return list, nil
	}

	pattern := "%" + strings.ToLower(q) + "%"
	if err := r.preload(r.db.WithContext(ctx)).
		Where("LOWER(name) LIKE ?", pattern).
		Order("updated_at DESC").
		Limit(20).
		Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (r *gormRepo) ListByUser(ctx context.Context, userID uint) ([]playlist.Playlist, error) {
	var list []playlist.Playlist
	if err := r.preload(r.db.WithContext(ctx)).Where("created_by_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormRepo) Update(ctx context.Context, p *playlist.Playlist) error {
	return r.db.WithContext(ctx).
		Model(&playlist.Playlist{}).
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"name":        p.Name,
			"description": p.Description,
			"is_public":   p.IsPublic,
		}).Error
}

func (r *gormRepo) ReplaceItems(ctx context.Context, playlistID uint, items []playlist.PlaylistItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("playlist_id = ?", playlistID).Delete(&playlist.PlaylistItem{}).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		for i := range items {
			items[i].PlaylistID = playlistID
			items[i].Position = uint(i + 1)
		}

		return tx.Create(&items).Error
	})
}

func (r *gormRepo) AddItems(ctx context.Context, playlistID uint, items []playlist.PlaylistItem) error {
	if len(items) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxPos uint
		if err := tx.Model(&playlist.PlaylistItem{}).
			Select("COALESCE(MAX(position), 0)").
			Where("playlist_id = ?", playlistID).
			Scan(&maxPos).Error; err != nil {
			return err
		}

		for i := range items {
			maxPos++
			items[i].PlaylistID = playlistID
			items[i].Position = maxPos
		}

		return tx.Create(&items).Error
	})
}

func (r *gormRepo) UpdateItem(ctx context.Context, itemID uint, defaultCoverID *string, note *string) error {
	updates := map[string]interface{}{}
	if defaultCoverID != nil {
		updates["default_cover_id"] = *defaultCoverID
	}
	if note != nil {
		updates["note"] = *note
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&playlist.PlaylistItem{}).
		Where("id = ?", itemID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *gormRepo) ReorderItems(ctx context.Context, playlistID uint, orders []playlist.ItemOrder) error {
	if len(orders) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, order := range orders {
			tmp := uint(1000000 + i + 1)
			if err := tx.Model(&playlist.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", order.ItemID, playlistID).
				Update("position", tmp).Error; err != nil {
				return err
			}
		}

		for _, order := range orders {
			if err := tx.Model(&playlist.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", order.ItemID, playlistID).
				Update("position", order.Position).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gormRepo) DeleteItem(ctx context.Context, playlistID uint, itemID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("id = ? AND playlist_id = ?", itemID, playlistID).Delete(&playlist.PlaylistItem{}).Error; err != nil {
			return err
		}

		var items []playlist.PlaylistItem
		if err := tx.Where("playlist_id = ?", playlistID).Order("position ASC").Find(&items).Error; err != nil {
			return err
		}

		for i := range items {
			tmp := uint(1000000 + i + 1)
			if err := tx.Model(&playlist.PlaylistItem{}).
				Where("id = ?", items[i].ID).
				Update("position", tmp).Error; err != nil {
				return err
			}
		}

		for i := range items {
			if err := tx.Model(&playlist.PlaylistItem{}).
				Where("id = ?", items[i].ID).
				Update("position", uint(i+1)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gormRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&playlist.Playlist{}, "id = ?", id).Error
}
