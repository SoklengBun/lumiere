package service

import (
	lyricsmodel "lumiere/internal/lyrics"
	"testing"
)

func TestIsValidDefaultCoverID(t *testing.T) {
	lyrics := lyricsmodel.Lyrics{
		VideoID: "original-video",
		Covers: []lyricsmodel.LyricCover{
			{CoverID: "cover-video"},
		},
	}

	tests := []struct {
		name           string
		defaultCoverID string
		want           bool
	}{
		{name: "original video", defaultCoverID: "original-video", want: true},
		{name: "cover video", defaultCoverID: "cover-video", want: true},
		{name: "unknown video", defaultCoverID: "unknown-video", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidDefaultCoverID(lyrics, tt.defaultCoverID); got != tt.want {
				t.Fatalf("isValidDefaultCoverID() = %v, want %v", got, tt.want)
			}
		})
	}
}
