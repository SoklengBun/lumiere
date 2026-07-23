package service

import (
	"context"
	"testing"

	lyricsmodel "lumiere/internal/lyrics"
	lyricsrepo "lumiere/internal/lyrics/repository"
)

type artistLyricsRepoStub struct {
	lyricsrepo.LyricsRepo
	artistID uint
	list     []lyricsmodel.Lyrics
}

func (s *artistLyricsRepoStub) ListByArtistID(_ context.Context, artistID uint) ([]lyricsmodel.Lyrics, error) {
	s.artistID = artistID
	return s.list, nil
}

func TestListByArtistIDDelegatesToRepository(t *testing.T) {
	stub := &artistLyricsRepoStub{
		list: []lyricsmodel.Lyrics{{Title: "Test Song"}},
	}
	svc := New(stub)

	got, err := svc.ListByArtistID(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListByArtistID() error = %v", err)
	}
	if stub.artistID != 42 {
		t.Fatalf("repository artist ID = %d, want 42", stub.artistID)
	}
	if len(got) != 1 || got[0].Title != "Test Song" {
		t.Fatalf("ListByArtistID() = %#v, want one Test Song result", got)
	}
}
