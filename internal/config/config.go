package config

import (
	"errors"
	"log"
	"os"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBName    string
	DBUser    string
	DBPass    string
	DBSSLMode string
	AppHost   string
	AppPort   string
	JWTSecret string
}

func NewFromEnv() (*Config, error) {
	get := func(keys ...string) string {
		for _, k := range keys {
			if v := os.Getenv(k); v != "" {
				return v
			}
		}
		return ""
	}

	cfg := &Config{
		DBHost:    get("DB_HOST", "DB.HOST"),
		DBPort:    get("DB_PORT", "DB.PORT"),
		DBName:    get("DB_NAME", "DB.NAME"),
		DBUser:    get("DB_USER", "DB.USER"),
		DBPass:    get("DB_PASS", "DB.PASS"),
		DBSSLMode: get("DB_SSLMODE", "DB.SSLMODE"),
		AppHost:   get("APP_HOST", "APP.HOST"),
		AppPort:   get("APP_PORT", "APP.PORT"),
		JWTSecret: get("JWT_SECRET", "JWT.SECRET"),
	}

	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPass == "" || cfg.AppHost == "" || cfg.JWTSecret == "" {
		return nil, errors.New("missing required environment variables")
	}

	log.Println("connection success")

	return cfg, nil
}
