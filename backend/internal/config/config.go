package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr           string
	PostgresDSN        string
	OpenAIAPIKey       string
	OpenAIModel        string
	ImagesDir          string
	CORSOrigins        []string
	DemoMode           bool
	SerpAPIKey         string
	MigrateOnStart     bool
	SessionSecret      string
	CookieSecure       bool
	FrontendURL        string
	APIPublicURL       string
	GoogleClientID     string
	GoogleClientSecret string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPassword       string
	SMTPFrom           string
	WorkerConcurrency  int
	JobMaxAttempts     int
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
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		PostgresDSN:        dsn,
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:        getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		ImagesDir:          getEnv("IMAGES_DIR", "images"),
		CORSOrigins:        origins,
		DemoMode:           getEnvBool("DEMO_MODE", false),
		SerpAPIKey:         os.Getenv("SERPAPI_KEY"),
		MigrateOnStart:     getEnvBool("MIGRATE_ON_START", true),
		SessionSecret:      getEnv("SESSION_SECRET", "dev-session-secret-change-me"),
		CookieSecure:       getEnvBool("COOKIE_SECURE", false),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		APIPublicURL:       getEnv("API_PUBLIC_URL", "http://localhost:8080"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:           getEnv("SMTP_FROM", "LOOKBOOK <noreply@lookbook.local>"),
		WorkerConcurrency:  getEnvInt("WORKER_CONCURRENCY", 4),
		JobMaxAttempts:     getEnvInt("JOB_MAX_ATTEMPTS", 3),
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

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
