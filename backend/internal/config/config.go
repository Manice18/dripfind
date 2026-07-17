package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr       string
	PostgresDSN    string
	OpenAIAPIKey   string
	OpenAIModel    string
	ImagesDir      string
	CORSOrigins    []string
	DemoMode       bool
	SerpAPIKey     string
	MigrateOnStart bool
}

func Load() (*Config, error) {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "outfit")
	pass := getEnv("POSTGRES_PASSWORD", "outfit")
	db := getEnv("POSTGRES_DB", "outfitfinder")
	ssl := getEnv("POSTGRES_SSLMODE", "disable")

	dsn := getEnv("DATABASE_URL", fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, db, ssl,
	))

	origins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	cfg := &Config{
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		PostgresDSN:    dsn,
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:    getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		ImagesDir:      getEnv("IMAGES_DIR", "images"),
		CORSOrigins:    origins,
		DemoMode:       getEnvBool("DEMO_MODE", false),
		SerpAPIKey:     os.Getenv("SERPAPI_KEY"),
		MigrateOnStart: getEnvBool("MIGRATE_ON_START", true),
	}

	if cfg.OpenAIAPIKey == "" {
		cfg.DemoMode = true
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
