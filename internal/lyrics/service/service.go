package service

import (
	"context"
	lyricsmodel "lumiere/internal/lyrics"
	lyricsrepo "lumiere/internal/lyrics/repository"
	"strings"
)

type Service struct {
	repo lyricsrepo.LyricsRepo
}

func New(repo lyricsrepo.LyricsRepo) *Service { return &Service{repo: repo} }

func (s *Service) Add(ctx context.Context, l *lyricsmodel.Lyrics) (*lyricsmodel.Lyrics, error) {
	if err := s.repo.Create(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*lyricsmodel.Lyrics, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByVideoID(ctx context.Context, videoID string) (*lyricsmodel.Lyrics, error) {
	videoID = strings.TrimSpace(videoID)
	return s.repo.GetByVideoID(ctx, videoID)
}

func (s *Service) List(ctx context.Context, page int, offset int) ([]lyricsmodel.Lyrics, int64, error) {
	return s.repo.List(ctx, page, offset)
}

func (s *Service) ListRandom(ctx context.Context, limit int) ([]lyricsmodel.Lyrics, error) {
	return s.repo.ListRandom(ctx, limit)
}

func (s *Service) ListByUser(ctx context.Context, userID uint) ([]lyricsmodel.Lyrics, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Search(ctx context.Context, q string) ([]lyricsmodel.Lyrics, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []lyricsmodel.Lyrics{}, nil
	}
	return s.repo.Search(ctx, q)
}

func (s *Service) Update(ctx context.Context, l *lyricsmodel.Lyrics) (*lyricsmodel.Lyrics, error) {
	if err := s.repo.Update(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}
