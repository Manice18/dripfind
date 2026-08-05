package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/manice18/dripfind/backend/internal/config"
	"github.com/manice18/dripfind/backend/internal/database"
	"github.com/manice18/dripfind/backend/internal/history"
	imagedl "github.com/manice18/dripfind/backend/internal/image"
	"github.com/manice18/dripfind/backend/internal/jobs"
	"github.com/manice18/dripfind/backend/internal/pinterest"
	"github.com/manice18/dripfind/backend/internal/pipeline"
	"github.com/manice18/dripfind/backend/internal/ranking"
	"github.com/manice18/dripfind/backend/internal/search"
	"github.com/manice18/dripfind/backend/internal/storage"
	"github.com/manice18/dripfind/backend/internal/vision"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Error("database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	imageStore, err := buildStorage(context.Background(), cfg)
	if err != nil {
		log.Error("storage", "error", err)
		os.Exit(1)
	}
	log.Info("storage ready", "backend", strings.ToLower(cfg.StorageBackend))

	var analyzer vision.Analyzer
	if cfg.DemoMode || cfg.OpenAIAPIKey == "" {
		log.Warn("running in demo vision mode (set OPENAI_API_KEY for real analysis)")
		analyzer = vision.NewDemoAnalyzer()
	} else {
		analyzer = vision.NewOpenAIAnalyzer(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	}

	if err := search.ConfigureScrape(
		cfg.ScrapeMaxConcurrent,
		time.Duration(cfg.ScrapeMinIntervalMS)*time.Millisecond,
		cfg.ScrapeProxyURL,
	); err != nil {
		log.Error("scrape config", "error", err)
		os.Exit(1)
	}
	if cfg.ScrapeProxyURL != "" {
		log.Info("scrape proxy enabled")
	}
	log.Info("scrape gate",
		"max_concurrent", cfg.ScrapeMaxConcurrent,
		"min_interval_ms", cfg.ScrapeMinIntervalMS,
		"breaker_threshold", cfg.ScrapeBreakerThreshold,
		"breaker_cooldown_sec", cfg.ScrapeBreakerCooldownS,
	)

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

	engine := search.NewEngine(log, providers...)
	engine.SetBreakerConfig(cfg.ScrapeBreakerThreshold, time.Duration(cfg.ScrapeBreakerCooldownS)*time.Second)

	outfitStore := history.NewStore(pool)
	pipe := &pipeline.Pipeline{
		Store:     outfitStore,
		Extractor: pinterest.NewExtractor(),
		Download:  imagedl.NewDownloader(),
		Storage:   imageStore,
		Vision:    analyzer,
		Search:    engine,
		Ranker:    ranking.NewScoreRanker(),
		Log:       log,
	}

	worker := &jobs.Worker{
		Jobs:        jobs.NewStore(pool, cfg.JobMaxAttempts),
		Outfits:     outfitStore,
		Pipeline:    pipe,
		Log:         log,
		Concurrency: cfg.WorkerConcurrency,
	}

	log.Info("worker process ready",
		"concurrency", cfg.WorkerConcurrency,
		"job_max_attempts", cfg.JobMaxAttempts,
		"demo_mode", cfg.DemoMode,
	)
	worker.Run(ctx)
}

func buildStorage(ctx context.Context, cfg *config.Config) (storage.ImageStorage, error) {
	imagesDir := cfg.ImagesDir
	if !filepath.IsAbs(imagesDir) {
		imagesDir = filepath.Join(findBackendRoot(), imagesDir)
	}
	return storage.BackendFromConfig(ctx, cfg.StorageBackend, storage.LocalConfig{
		Dir:        imagesDir,
		PublicBase: "/images",
	}, storage.S3Config{
		Endpoint:       cfg.S3Endpoint,
		Region:         cfg.S3Region,
		Bucket:         cfg.S3Bucket,
		AccessKey:      cfg.S3AccessKeyID,
		SecretKey:      cfg.S3SecretAccessKey,
		PublicBaseURL:  cfg.S3PublicBaseURL,
		ForcePathStyle: cfg.S3ForcePathStyle,
	})
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
