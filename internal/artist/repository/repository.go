package repository

import (
	"context"
	"lumiere/internal/artist"
)

type ArtistRepo interface {
	Create(ctx context.Context, a *artist.Artist) error
	Update(ctx context.Context, a *artist.Artist) error
	GetByID(ctx context.Context, id uint) (*artist.Artist, error)
	FindByIDs(ctx context.Context, ids []uint) ([]artist.Artist, error)
	FindByName(ctx context.Context, q string) ([]artist.Artist, error)
	List(ctx context.Context) ([]artist.Artist, error)
}
