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
	imagedl "github.com/manice18/outfit_finder/backend/internal/image"
	"github.com/manice18/outfit_finder/backend/internal/pinterest"
	"github.com/manice18/outfit_finder/backend/internal/pipeline"
	"github.com/manice18/outfit_finder/backend/internal/ranking"
	"github.com/manice18/outfit_finder/backend/internal/search"
	"github.com/manice18/outfit_finder/backend/internal/storage"
	"github.com/manice18/outfit_finder/backend/internal/vision"
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

	var analyzer vision.Analyzer
	if cfg.DemoMode || cfg.OpenAIAPIKey == "" {
		log.Warn("running in demo vision mode (set OPENAI_API_KEY for real analysis)")
		analyzer = vision.NewDemoAnalyzer()
	} else {
		analyzer = vision.NewOpenAIAnalyzer(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	}

	providers := []search.Provider{
		search.NewMyntraProvider(),
		search.NewAjioProvider(),
		search.NewFlipkartProvider(),
		search.NewBewakoofProvider(),
		search.NewHMProvider(),
	}
	providers = append(providers, search.HomegrownShopifyBrands()...)
	log.Info("search providers registered", "count", len(providers))
	if cfg.SerpAPIKey != "" {
		providers = append(providers, search.NewSerpShoppingProvider(cfg.SerpAPIKey))
		log.Info("serp shopping provider enabled")
	}

	store := history.NewStore(pool)
	pipe := &pipeline.Pipeline{
		Store:     store,
		Extractor: pinterest.NewExtractor(),
		Download:  imagedl.NewDownloader(),
		Storage:   localStore,
		Vision:    analyzer,
		Search:    search.NewEngine(log, providers...),
		Ranker:    ranking.NewScoreRanker(),
		Log:       log,
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
		Store:    store,
		Pipeline: pipe,
		Auth:     authSvc,
		Log:      log,
		Origins:  cfg.CORSOrigins,
		Images:   http.FileServer(http.Dir(imagesDir)),
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
