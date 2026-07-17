package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/manice18/outfit_finder/backend/internal/history"
	imagedl "github.com/manice18/outfit_finder/backend/internal/image"
	"github.com/manice18/outfit_finder/backend/internal/models"
	"github.com/manice18/outfit_finder/backend/internal/pinterest"
	"github.com/manice18/outfit_finder/backend/internal/ranking"
	"github.com/manice18/outfit_finder/backend/internal/search"
	"github.com/manice18/outfit_finder/backend/internal/storage"
	"github.com/manice18/outfit_finder/backend/internal/vision"
)

type Pipeline struct {
	Store     *history.Store
	Extractor *pinterest.Extractor
	Download  *imagedl.Downloader
	Storage   storage.ImageStorage
	Vision    vision.Analyzer
	Search    *search.Engine
	Ranker    ranking.Ranker
	Log       *slog.Logger
}

func (p *Pipeline) Start(ctx context.Context, outfitID uuid.UUID, sourceURL string) {
	go func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		if err := p.run(runCtx, outfitID, sourceURL); err != nil {
			p.Log.Error("pipeline failed", "outfit_id", outfitID, "error", err)
			_ = p.Store.MarkFailed(context.Background(), outfitID, err.Error())
		}
	}()
	_ = ctx
}

func (p *Pipeline) run(ctx context.Context, outfitID uuid.UUID, sourceURL string) error {
	log := p.Log.With("outfit_id", outfitID)
	start := time.Now()

	if err := p.Store.MarkProcessing(ctx, outfitID); err != nil {
		return err
	}

	log.Info("extracting image url", "source_url", sourceURL)
	imageURL, err := p.Extractor.ExtractImageURL(sourceURL)
	if err != nil {
		return fmt.Errorf("extract image: %w", err)
	}

	log.Info("downloading image", "image_url", imageURL)
	dl, err := p.Download.Download(imageURL)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}

	rel, err := p.Storage.Save(dl.Data, dl.ContentType)
	if err != nil {
		return fmt.Errorf("save image: %w", err)
	}
	abs := p.Storage.AbsPath(rel)
	log.Info("image saved", "path", rel)

	visionStart := time.Now()
	analyzed, err := p.Vision.Analyze(ctx, abs)
	if err != nil {
		return fmt.Errorf("vision: %w", err)
	}
	log.Info("vision complete", "latency_ms", time.Since(visionStart).Milliseconds(), "items", len(analyzed.Items))

	outfit := &models.Outfit{
		ID:        outfitID,
		ImagePath: filepath.ToSlash(rel),
		SourceURL: sourceURL,
		Style:     analyzed.Style,
		Gender:    analyzed.Gender,
		Season:    analyzed.Season,
		Occasion:  analyzed.Occasion,
		Status:    models.StatusCompleted,
	}

	rankStart := time.Now()
	for _, vi := range analyzed.Items {
		item := models.ClothingItem{
			Category:    vi.Category,
			Color:       vi.Color,
			Material:    vi.Material,
			Fit:         vi.Fit,
			Pattern:     vi.Pattern,
			Confidence:  vi.Confidence,
			SearchQuery: search.BuildQueryFromVision(vi, analyzed.Gender),
		}

		result := p.Search.SearchItem(ctx, item, analyzed.Gender)
		ranked := p.Ranker.Rank(item, result.Products, analyzed.Gender)
		item.Products = ranked
		outfit.Items = append(outfit.Items, item)
		log.Info("item search ranked",
			"category", item.Category,
			"query", item.SearchQuery,
			"raw", len(result.Products),
			"kept", len(ranked),
			"provider_errors", len(result.Errs),
		)
	}
	log.Info("ranking complete", "latency_ms", time.Since(rankStart).Milliseconds())

	if err := p.Store.SaveAnalysis(ctx, outfit); err != nil {
		return fmt.Errorf("persist: %w", err)
	}

	log.Info("pipeline complete", "total_ms", time.Since(start).Milliseconds())
	return nil
}
