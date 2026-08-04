package pipeline

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/manice18/dripfind/backend/internal/history"
	imagedl "github.com/manice18/dripfind/backend/internal/image"
	"github.com/manice18/dripfind/backend/internal/models"
	"github.com/manice18/dripfind/backend/internal/pinterest"
	"github.com/manice18/dripfind/backend/internal/ranking"
	"github.com/manice18/dripfind/backend/internal/search"
	"github.com/manice18/dripfind/backend/internal/storage"
	"github.com/manice18/dripfind/backend/internal/vision"
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

// RunURL extracts and analyzes a Pinterest (or similar) URL. Called by the worker.
func (p *Pipeline) RunURL(ctx context.Context, outfitID uuid.UUID, sourceURL string) error {
	log := p.Log.With("outfit_id", outfitID)

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

	return p.finishFromImage(ctx, outfitID, sourceURL, dl.Data, dl.ContentType)
}

// RunUpload analyzes an image already written to storage (key in the job payload).
func (p *Pipeline) RunUpload(ctx context.Context, outfitID uuid.UUID, imageKey, sourceLabel string) error {
	if err := p.Store.MarkProcessing(ctx, outfitID); err != nil {
		return err
	}
	if sourceLabel == "" {
		sourceLabel = "upload"
	}
	return p.finishFromStored(ctx, outfitID, sourceLabel, imageKey)
}

func (p *Pipeline) finishFromImage(ctx context.Context, outfitID uuid.UUID, sourceURL string, data []byte, contentType string) error {
	key, err := p.Storage.Save(ctx, data, contentType)
	if err != nil {
		return fmt.Errorf("save image: %w", err)
	}
	p.Log.Info("image saved to object storage", "outfit_id", outfitID, "key", key, "bytes", len(data), "url", p.Storage.PublicURL(key))
	return p.finishFromBytes(ctx, outfitID, sourceURL, key, data, contentType)
}

func (p *Pipeline) finishFromStored(ctx context.Context, outfitID uuid.UUID, sourceURL, key string) error {
	rc, err := p.Storage.Open(ctx, key)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read image: %w", err)
	}
	return p.finishFromBytes(ctx, outfitID, sourceURL, key, data, storage.ContentTypeFromKey(key))
}

func (p *Pipeline) finishFromBytes(ctx context.Context, outfitID uuid.UUID, sourceURL, key string, data []byte, contentType string) error {
	log := p.Log.With("outfit_id", outfitID)
	start := time.Now()
	log.Info("image ready", "key", key)

	visionStart := time.Now()
	analyzed, err := p.Vision.Analyze(ctx, data, contentType)
	if err != nil {
		return fmt.Errorf("vision: %w", err)
	}
	log.Info("vision complete", "latency_ms", time.Since(visionStart).Milliseconds(), "items", len(analyzed.Items))

	outfit := &models.Outfit{
		ID:        outfitID,
		ImagePath: path.Clean(key),
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
			SearchQuery: search.BuildQueryFromVision(vi, analyzed.Gender, analyzed.Occasion),
		}

		result := p.Search.SearchItem(ctx, item, analyzed.Gender)
		ranked := p.Ranker.Rank(item, result.Products, analyzed.Gender, analyzed.Occasion)
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
