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
		DBHost:    get("DB_HOST"),
		DBPort:    get("DB_PORT"),
		DBName:    get("DB_NAME"),
		DBUser:    get("DB_USER"),
		DBPass:    get("DB_PASS"),
		DBSSLMode: get("DB_SSLMODE"),
		AppHost:   get("APP_HOST"),
		AppPort:   get("APP_PORT"),
		JWTSecret: get("JWT_SECRET"),
	}

	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPass == "" || cfg.AppHost == "" || cfg.JWTSecret == "" {
		return nil, errors.New("missing required environment variables")
	}

	log.Println("connection success")

	return cfg, nil
}
