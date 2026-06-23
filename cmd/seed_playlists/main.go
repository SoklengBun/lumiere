package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"lumiere/internal/config"
	"lumiere/internal/database"
	"lumiere/internal/lyrics"
	"lumiere/internal/playlist"
)

func main() {
	playlistCount := flag.Int("count", 3, "number of playlists to create")
	itemsPerPlaylist := flag.Int("items", 5, "number of songs per playlist")
	wipe := flag.Bool("wipe", false, "delete existing playlists before seeding")
	flag.Parse()

	if *playlistCount <= 0 {
		log.Fatal("count must be > 0")
	}
	if *itemsPerPlaylist <= 0 {
		log.Fatal("items must be > 0")
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if *wipe {
			if err := tx.Exec("TRUNCATE TABLE playlist_items, playlists RESTART IDENTITY CASCADE").Error; err != nil {
				return err
			}
		}

		required := (*playlistCount) * (*itemsPerPlaylist)
		var songs []lyrics.Lyrics
		if err := tx.Order("RANDOM()").Limit(required).Find(&songs).Error; err != nil {
			return err
		}
		if len(songs) == 0 {
			return fmt.Errorf("no songs found in lyrics table")
		}

		names := []string{
			"Chill J-Pop Nights",
			"Anime Love Songs",
			"Vocaloid and Friends",
			"City Pop and Ballads",
			"Late Night Lyrics",
		}

		idx := 0
		for i := 0; i < *playlistCount; i++ {
			p := &playlist.Playlist{
				Name:        names[i%len(names)],
				Description: "Auto-generated sample playlist",
				IsPublic:    true,
				CreatedByID: 0,
				Items:       make([]playlist.PlaylistItem, 0, *itemsPerPlaylist),
			}

			for j := 0; j < *itemsPerPlaylist; j++ {
				if idx >= len(songs) {
					break
				}
				p.Items = append(p.Items, playlist.PlaylistItem{
					LyricsID: songs[idx].ID,
					Position: uint(j + 1),
					Note:     strings.TrimSpace(songs[idx].Title),
				})
				idx++
			}

			if len(p.Items) == 0 {
				break
			}

			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(p).Error; err != nil {
				return err
			}

			fmt.Printf("created playlist id=%d name=%q items=%d\n", p.ID, p.Name, len(p.Items))
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
