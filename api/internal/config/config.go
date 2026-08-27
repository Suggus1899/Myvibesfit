package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               string
	Env                string
	DatabaseURL        string
	JWTAccessSecret    string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("ENV", "development"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTAccessSecret: os.Getenv("JWT_ACCESS_SECRET"),
	}

	// En development, sin config explicita, se abre a los puertos locales
	// tipicos (Flutter web, panel Next.js). En produccion hay que listarlos.
	origins := getEnv("CORS_ALLOWED_ORIGINS", "")
	if origins == "" && cfg.Env == "development" {
		origins = "http://localhost:5555,http://127.0.0.1:5555,http://localhost:3000,http://127.0.0.1:3000"
	}
	if origins != "" {
		cfg.CORSAllowedOrigins = strings.Split(origins, ",")
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
