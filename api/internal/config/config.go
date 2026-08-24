package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	Env             string
	DatabaseURL     string
	JWTAccessSecret string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("ENV", "development"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTAccessSecret: os.Getenv("JWT_ACCESS_SECRET"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTAccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}

	accessMinutes, err := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL_MINUTES: %w", err)
	}
	cfg.JWTAccessTTL = time.Duration(accessMinutes) * time.Minute

	refreshDays, err := strconv.Atoi(getEnv("JWT_REFRESH_TTL_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL_DAYS: %w", err)
	}
	cfg.JWTRefreshTTL = time.Duration(refreshDays) * 24 * time.Hour

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
