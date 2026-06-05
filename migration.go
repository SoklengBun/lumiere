package main

import (
	"log"

	"github.com/joho/godotenv"

	"lumiere/internal/artist"
	"lumiere/internal/config"
	"lumiere/internal/database"
	"lumiere/internal/lyrics"
	"lumiere/internal/models"
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
		&lyrics.LyricTitle{},
		&lyrics.LyricContent{},
		&lyrics.LyricReference{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}
}
