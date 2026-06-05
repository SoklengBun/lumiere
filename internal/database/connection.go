package database

import (
	"fmt"

	"os"
	"strconv"
	"time"

	"lumiere/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	host := cfg.DBHost
	port := cfg.DBPort
	name := cfg.DBName
	user := cfg.DBUser
	pwd := cfg.DBPass
	sslmode := cfg.DBSSLMode

	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, pwd, name, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	getenvInt := func(key string, def int) int {
		if v := os.Getenv(key); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
		return def
	}

	maxOpen := getenvInt("DB_MAX_OPEN_CONNS", 25)
	maxIdle := getenvInt("DB_MAX_IDLE_CONNS", 25)
	connMaxLifetimeMins := getenvInt("DB_CONN_MAX_LIFETIME_MIN", 15)

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetimeMins) * time.Minute)

	return db, nil
}
