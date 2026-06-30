package repository

import (
	"context"
	"lumiere/internal/playlist"
)

type PlaylistRepo interface {
	Create(ctx context.Context, p *playlist.Playlist) error
	GetByID(ctx context.Context, id uint) (*playlist.Playlist, error)
	GetItemByID(ctx context.Context, itemID uint) (*playlist.PlaylistItem, error)
	ListPublic(ctx context.Context) ([]playlist.Playlist, error)
	ListByUser(ctx context.Context, userID uint) ([]playlist.Playlist, error)
	SearchByName(ctx context.Context, q string) ([]playlist.Playlist, error)
	Update(ctx context.Context, p *playlist.Playlist) error
	ReplaceItems(ctx context.Context, playlistID uint, items []playlist.PlaylistItem) error
	AddItems(ctx context.Context, playlistID uint, items []playlist.PlaylistItem) error
	UpdateItem(ctx context.Context, itemID uint, defaultCoverID *string, note *string) error
	ReorderItems(ctx context.Context, playlistID uint, orders []playlist.ItemOrder) error
	DeleteItem(ctx context.Context, playlistID uint, itemID uint) error
	Delete(ctx context.Context, id uint) error
}
