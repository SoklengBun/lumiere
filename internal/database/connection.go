package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB.HOST")
	port := os.Getenv("DB.PORT")
	name := os.Getenv("DB.NAME")
	user := os.Getenv("DB.USER")
	pwd := os.Getenv("DB.PASS")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", host, user, pwd, name, port)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
