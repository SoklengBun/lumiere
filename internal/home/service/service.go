package service

import (
	"context"
	artistmodel "lumiere/internal/artist"
	artistsvc "lumiere/internal/artist/service"
	lyricsmodel "lumiere/internal/lyrics"
	lyricssvc "lumiere/internal/lyrics/service"
	playlistmodel "lumiere/internal/playlist"
	playlistsvc "lumiere/internal/playlist/service"
)

const (
	defaultSongsLimit     = 10
	defaultPlaylistsLimit = 5
	defaultPlaylistSongs  = 10
	defaultArtistsLimit   = 5
)

type Payload struct {
	Songs     []lyricsmodel.Lyrics     `json:"songs"`
	Playlists []playlistmodel.Playlist `json:"playlists"`
	Artists   []artistmodel.Artist     `json:"artists"`
}

type Service struct {
	lyricsSvc   *lyricssvc.Service
	playlistSvc *playlistsvc.Service
	artistSvc   *artistsvc.Service
}

func New(lyricsSvc *lyricssvc.Service, playlistSvc *playlistsvc.Service, artistSvc *artistsvc.Service) *Service {
	return &Service{lyricsSvc: lyricsSvc, playlistSvc: playlistSvc, artistSvc: artistSvc}
}

func (s *Service) Get(ctx context.Context) (*Payload, error) {
	songs, err := s.lyricsSvc.ListRandom(ctx, defaultSongsLimit)
	if err != nil {
		return nil, err
	}

	playlists, err := s.playlistSvc.ListPublic(ctx)
	if err != nil {
		return nil, err
	}

	if len(playlists) > defaultPlaylistsLimit {
		playlists = playlists[:defaultPlaylistsLimit]
	}

	for i := range playlists {
		if len(playlists[i].Items) > defaultPlaylistSongs {
			playlists[i].Items = playlists[i].Items[:defaultPlaylistSongs]
		}
	}

	artists, err := s.artistSvc.ListByRecentLyrics(ctx, defaultArtistsLimit)
	if err != nil {
		return nil, err
	}

	return &Payload{
		Songs:     songs,
		Playlists: playlists,
		Artists:   artists,
	}, nil
}
