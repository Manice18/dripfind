package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/manice18/outfit_finder/backend/internal/api"
	"github.com/manice18/outfit_finder/backend/internal/auth"
	"github.com/manice18/outfit_finder/backend/internal/config"
	"github.com/manice18/outfit_finder/backend/internal/database"
	"github.com/manice18/outfit_finder/backend/internal/history"
	"github.com/manice18/outfit_finder/backend/internal/jobs"
	"github.com/manice18/outfit_finder/backend/internal/storage"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Error("database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.MigrateOnStart {
		migDir := findMigrations()
		if err := database.RunMigrations(ctx, pool, migDir); err != nil {
			log.Error("migrate", "error", err)
			os.Exit(1)
		}
		log.Info("migrations applied", "dir", migDir)
	}

	imagesDir := cfg.ImagesDir
	if !filepath.IsAbs(imagesDir) {
		imagesDir = filepath.Join(findBackendRoot(), imagesDir)
	}
	localStore, err := storage.NewLocalStorage(imagesDir)
	if err != nil {
		log.Error("storage", "error", err)
		os.Exit(1)
	}

	var mailer auth.Mailer = auth.LogMailer{Log: log}
	if cfg.SMTPHost != "" {
		mailer = auth.SMTPMailer{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			User:     cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			Log:      log,
		}
	}

	authSvc := &auth.Service{
		Store:              auth.NewStore(pool),
		Mailer:             mailer,
		Log:                log,
		SessionSecret:      cfg.SessionSecret,
		CookieSecure:       cfg.CookieSecure,
		FrontendURL:        cfg.FrontendURL,
		APIPublicURL:       cfg.APIPublicURL,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
	}

	srv := &api.Server{
		Store:          history.NewStore(pool),
		Jobs:           jobs.NewStore(pool, cfg.JobMaxAttempts),
		Storage:        localStore,
		Auth:           authSvc,
		Log:            log,
		Origins:        cfg.CORSOrigins,
		Images:         http.FileServer(http.Dir(imagesDir)),
		JobMaxAttempts: cfg.JobMaxAttempts,
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr, "demo_mode", cfg.DemoMode)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func findBackendRoot() string {
	candidates := []string{".", "..", filepath.Join("..", "backend")}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "migrations")); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	abs, _ := filepath.Abs(".")
	return abs
}

func findMigrations() string {
	return filepath.Join(findBackendRoot(), "migrations")
}
