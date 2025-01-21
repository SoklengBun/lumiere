package main

import (
	"log"

	"lumiere/internal/database"
	"lumiere/internal/models"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal(err.Error())
	}
}
