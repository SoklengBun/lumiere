package repository

import (
	"context"
	"lumiere/internal/lyrics"
)

type LyricsRepo interface {
	Create(ctx context.Context, l *lyrics.Lyrics) error
	GetByID(ctx context.Context, id uint) (*lyrics.Lyrics, error)
	GetByVideoID(ctx context.Context, videoID string) (*lyrics.Lyrics, error)
	List(ctx context.Context, page int, offset int) ([]lyrics.Lyrics, int64, error)
	ListRandom(ctx context.Context, limit int) ([]lyrics.Lyrics, error)
	ListByUser(ctx context.Context, userID uint) ([]lyrics.Lyrics, error)
	Search(ctx context.Context, q string) ([]lyrics.Lyrics, error)
	Update(ctx context.Context, l *lyrics.Lyrics) error
}
