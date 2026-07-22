package service

import (
	"context"
	"errors"
	lyricsmodel "lumiere/internal/lyrics"
	lyricssvc "lumiere/internal/lyrics/service"
	"lumiere/internal/playlist"
	playlistrepo "lumiere/internal/playlist/repository"
	"strings"
)

type Service struct {
	repo      playlistrepo.PlaylistRepo
	lyricsSvc *lyricssvc.Service
}

func New(repo playlistrepo.PlaylistRepo, lyricsSvc *lyricssvc.Service) *Service {
	return &Service{repo: repo, lyricsSvc: lyricsSvc}
}

func (s *Service) Create(ctx context.Context, p *playlist.Playlist) (*playlist.Playlist, error) {
	if err := s.validateLyricsIDs(ctx, p.Items); err != nil {
		return nil, err
	}
	if err := validateUniqueLyricsIDs(p.Items, nil); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*playlist.Playlist, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListPublic(ctx context.Context) ([]playlist.Playlist, error) {
	return s.repo.ListPublic(ctx)
}

func (s *Service) ListByUser(ctx context.Context, userID uint) ([]playlist.Playlist, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Search(ctx context.Context, q string) ([]playlist.Playlist, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []playlist.Playlist{}, nil
	}
	return s.repo.SearchByName(ctx, q)
}

func (s *Service) Update(ctx context.Context, p *playlist.Playlist) (*playlist.Playlist, error) {
	if err := s.validateLyricsIDs(ctx, p.Items); err != nil {
		return nil, err
	}
	if err := validateUniqueLyricsIDs(p.Items, nil); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceItems(ctx, p.ID, p.Items); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) AddItems(ctx context.Context, playlistID uint, items []playlist.PlaylistItem) error {
	if err := s.validateLyricsIDs(ctx, items); err != nil {
		return err
	}
	current, err := s.repo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if err := validateUniqueLyricsIDs(items, current.Items); err != nil {
		return err
	}
	return s.repo.AddItems(ctx, playlistID, items)
}

func (s *Service) GetItem(ctx context.Context, itemID uint) (*playlist.PlaylistItem, error) {
	return s.repo.GetItemByID(ctx, itemID)
}

func (s *Service) UpdateItem(ctx context.Context, itemID uint, defaultCoverID *string, note *string) error {
	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	if defaultCoverID != nil {
		trimmedCoverID := strings.TrimSpace(*defaultCoverID)
		if trimmedCoverID != "" && !isValidDefaultCoverID(item.Lyrics, trimmedCoverID) {
			return errors.New("default cover ID is invalid")
		}
		defaultCoverID = &trimmedCoverID
	}

	if note != nil {
		trimmedNote := strings.TrimSpace(*note)
		note = &trimmedNote
	}

	return s.repo.UpdateItem(ctx, itemID, defaultCoverID, note)
}

func (s *Service) ReorderItems(ctx context.Context, playlistID uint, orders []playlist.ItemOrder) error {
	if len(orders) == 0 {
		return errors.New("missing item orders")
	}
	for _, o := range orders {
		if o.ItemID == 0 || o.Position == 0 {
			return errors.New("invalid item orders")
		}
	}
	return s.repo.ReorderItems(ctx, playlistID, orders)
}

func (s *Service) DeleteItem(ctx context.Context, playlistID uint, itemID uint) error {
	return s.repo.DeleteItem(ctx, playlistID, itemID)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) validateLyricsIDs(ctx context.Context, items []playlist.PlaylistItem) error {
	for i := range items {
		items[i].DefaultCoverID = strings.TrimSpace(items[i].DefaultCoverID)

		if items[i].LyricsID == 0 {
			return errors.New("invalid lyrics id")
		}
		lyrics, err := s.lyricsSvc.Get(ctx, items[i].LyricsID)
		if err != nil {
			return errors.New("one or more lyrics IDs are invalid")
		}
		if items[i].DefaultCoverID == "" {
			continue
		}

		if !isValidDefaultCoverID(*lyrics, items[i].DefaultCoverID) {
			return errors.New("one or more default cover IDs are invalid")
		}
	}
	return nil
}

func isValidDefaultCoverID(lyrics lyricsmodel.Lyrics, defaultCoverID string) bool {
	if lyrics.VideoID == defaultCoverID {
		return true
	}

	for _, cover := range lyrics.Covers {
		if cover.CoverID == defaultCoverID {
			return true
		}
	}

	return false
}

func validateUniqueLyricsIDs(items []playlist.PlaylistItem, existing []playlist.PlaylistItem) error {
	seen := make(map[uint]struct{}, len(items)+len(existing))

	for _, item := range existing {
		if item.LyricsID == 0 {
			continue
		}
		seen[item.LyricsID] = struct{}{}
	}

	for _, item := range items {
		if item.LyricsID == 0 {
			continue
		}
		if _, exists := seen[item.LyricsID]; exists {
			return errors.New("playlist already contains this song")
		}
		seen[item.LyricsID] = struct{}{}
	}

	return nil
}
