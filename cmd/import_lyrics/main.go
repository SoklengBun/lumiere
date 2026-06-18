package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"lumiere/internal/artist"
	"lumiere/internal/config"
	"lumiere/internal/database"
	"lumiere/internal/lyrics"
)

func main() {
	csvPath := flag.String("csv", "", "absolute path to csv file")
	truncate := flag.Bool("truncate", true, "truncate existing app data before importing")
	flag.Parse()

	if strings.TrimSpace(*csvPath) == "" {
		log.Fatal("missing required --csv")
	}

	_ = godotenv.Load()

	cfg, err := config.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	header, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}
	if len(header) < 5 {
		log.Fatal("invalid csv header")
	}

	artistCache := map[string]artist.Artist{}
	createdArtists := 0
	createdLyrics := 0

	err = db.Transaction(func(tx *gorm.DB) error {
		if *truncate {
			if err := truncateAll(tx); err != nil {
				return err
			}
		}

		for {
			rec, err := r.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
			if len(rec) < 5 {
				continue
			}

			title := strings.TrimSpace(rec[0])
			artistCell := strings.TrimSpace(rec[1])
			jp := strings.TrimSpace(rec[2])
			romaji := strings.TrimSpace(rec[3])
			youtube := strings.TrimSpace(rec[4])

			if title == "" {
				continue
			}

			artists, newlyCreated, err := resolveArtists(tx, artistCell, artistCache)
			if err != nil {
				return err
			}
			createdArtists += newlyCreated

			contents := make([]lyrics.LyricContent, 0, 2)
			if jp != "" {
				contents = append(contents, lyrics.LyricContent{Kind: "japanese", Content: jp})
			}
			if romaji != "" {
				contents = append(contents, lyrics.LyricContent{Kind: "romaji", Content: romaji})
			}

			songID := pickSongID(youtube)
			if songID == "" {
				continue
			}

			l := &lyrics.Lyrics{
				ID:      songID,
				Summary: title,
				Titles: []lyrics.LyricTitle{
					{Title: title, Normalized: strings.ToLower(title), Lang: "ja"},
				},
				Artists:  artists,
				Contents: contents,
			}

			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(l).Error; err != nil {
				return err
			}
			createdLyrics++
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("import completed: artists=%d lyrics=%d\n", createdArtists, createdLyrics)
}

func truncateAll(tx *gorm.DB) error {
	query := `TRUNCATE TABLE
		playlist_items,
		playlists,
		lyrics_artists,
		lyric_contents,
		lyric_titles,
		lyrics,
		artists,
		users
	RESTART IDENTITY CASCADE`
	return tx.Exec(query).Error
}

func resolveArtists(tx *gorm.DB, cell string, cache map[string]artist.Artist) ([]artist.Artist, int, error) {
	if strings.TrimSpace(cell) == "" {
		return nil, 0, nil
	}

	parts := strings.Split(cell, ",")
	result := make([]artist.Artist, 0, len(parts))
	created := 0

	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}

		norm := strings.ToLower(name)
		if cached, ok := cache[norm]; ok {
			result = append(result, cached)
			continue
		}

		var a artist.Artist
		err := tx.Where("LOWER(name) = ?", norm).First(&a).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a = artist.Artist{Name: name}
			if err := tx.Create(&a).Error; err != nil {
				return nil, 0, err
			}
			created++
		} else if err != nil {
			return nil, 0, err
		}

		cache[norm] = a
		result = append(result, a)
	}

	return result, created, nil
}

func pickSongID(rawCell string) string {
	return extractYouTubeID(rawCell)
}

func extractYouTubeID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")

	if strings.Contains(raw, "youtube.com/watch?v=") {
		parts := strings.SplitN(raw, "v=", 2)
		if len(parts) == 2 {
			idPart := strings.Split(parts[1], "&")[0]
			return strings.TrimSpace(idPart)
		}
	}

	if strings.Contains(raw, "youtu.be/") {
		parts := strings.SplitN(raw, "youtu.be/", 2)
		if len(parts) == 2 {
			idPart := strings.Split(parts[1], "?")[0]
			return strings.TrimSpace(idPart)
		}
	}

	if !strings.Contains(raw, "/") && !strings.Contains(raw, " ") {
		return raw
	}

	return ""
}
