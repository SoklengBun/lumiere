package main

import (
	"log"

	"github.com/joho/godotenv"

	"lumiere/internal/artist"
	"lumiere/internal/config"
	"lumiere/internal/database"
	"lumiere/internal/lyrics"
	"lumiere/internal/models"
	"lumiere/internal/playlist"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.NewFromEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.AutoMigrate(
		&models.User{},
		&artist.Artist{},
		&lyrics.Lyrics{},
		&lyrics.LyricCover{},
		&lyrics.LyricContent{},
		&playlist.Playlist{},
		&playlist.PlaylistItem{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}
}
