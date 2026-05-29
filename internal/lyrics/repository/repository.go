package repository

import (
	"context"
	"lumiere/internal/lyrics"
)

type LyricsRepo interface {
	Create(ctx context.Context, l *lyrics.Lyrics) error
	GetByID(ctx context.Context, id uint) (*lyrics.Lyrics, error)
	List(ctx context.Context) ([]lyrics.Lyrics, error)
	ListByUser(ctx context.Context, userID uint) ([]lyrics.Lyrics, error)
	Update(ctx context.Context, l *lyrics.Lyrics) error
}
