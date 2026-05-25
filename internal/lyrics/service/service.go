package service

import (
	"context"
	lyricsmodel "lumiere/internal/lyrics"
	lyricsrepo "lumiere/internal/lyrics/repository"
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

func (s *Service) List(ctx context.Context) ([]lyricsmodel.Lyrics, error) {
	return s.repo.List(ctx)
}
