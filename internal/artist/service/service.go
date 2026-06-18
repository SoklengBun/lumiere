package service

import (
	"context"
	"lumiere/internal/artist"
	artistrepo "lumiere/internal/artist/repository"
)

type Service struct{ repo artistrepo.ArtistRepo }

func New(repo artistrepo.ArtistRepo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, a *artist.Artist) error {
	return s.repo.Create(ctx, a)
}

func (s *Service) Update(ctx context.Context, a *artist.Artist) error {
	return s.repo.Update(ctx, a)
}

func (s *Service) GetByID(ctx context.Context, id uint) (*artist.Artist, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) FindByIDs(ctx context.Context, ids []uint) ([]artist.Artist, error) {
	return s.repo.FindByIDs(ctx, ids)
}

func (s *Service) FindByName(ctx context.Context, q string) ([]artist.Artist, error) {
	return s.repo.FindByName(ctx, q)
}

func (s *Service) List(ctx context.Context) ([]artist.Artist, error) {
	return s.repo.List(ctx)
}
