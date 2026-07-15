package config

import (
	"errors"
	"log"
	"os"
	"strconv"
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

	RedisEnabled bool
	RedisURL     string

	UpstashRedisRESTURL   string
	UpstashRedisRESTToken string
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

	upstashRedisRESTURL := get("UPSTASH_REDIS_REST_URL")
	upstashRedisRESTToken := get("UPSTASH_REDIS_REST_TOKEN")

	redisEnabled := upstashRedisRESTURL != "" && upstashRedisRESTToken != ""
	if raw := get("REDIS_ENABLED"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, errors.New("REDIS_ENABLED must be true or false")
		}
		redisEnabled = parsed
	}

	redisURL := get("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	cfg := &Config{
		DBHost:                get("DB_HOST"),
		DBPort:                get("DB_PORT"),
		DBName:                get("DB_NAME"),
		DBUser:                get("DB_USER"),
		DBPass:                get("DB_PASS"),
		DBSSLMode:             get("DB_SSLMODE"),
		AppHost:               get("APP_HOST"),
		AppPort:               get("APP_PORT"),
		JWTSecret:             get("JWT_SECRET"),
		RedisEnabled:          redisEnabled,
		RedisURL:              redisURL,
		UpstashRedisRESTURL:   upstashRedisRESTURL,
		UpstashRedisRESTToken: upstashRedisRESTToken,
	}

	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPass == "" || cfg.JWTSecret == "" {
		return nil, errors.New("missing required environment variables")
	}

	log.Println("connection success")

	return cfg, nil
}
