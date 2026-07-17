package search

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

type Provider interface {
	Name() string
	Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error)
}

type Engine struct {
	providers []Provider
	log       *slog.Logger
}

func NewEngine(log *slog.Logger, providers ...Provider) *Engine {
	return &Engine{providers: providers, log: log}
}

type itemResult struct {
	Products []models.Product
	Errs     []error
}

func (e *Engine) SearchItem(ctx context.Context, item models.ClothingItem, gender string) itemResult {
	type providerOut struct {
		name     string
		products []models.Product
		err      error
		ms       int64
	}

	ch := make(chan providerOut, len(e.providers))
	var wg sync.WaitGroup

	for _, p := range e.providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			start := time.Now()
			products, err := prov.Search(ctx, item, gender)
			ch <- providerOut{
				name:     prov.Name(),
				products: products,
				err:      err,
				ms:       time.Since(start).Milliseconds(),
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var out itemResult
	for r := range ch {
		e.log.Info("provider search finished",
			"provider", r.name,
			"category", item.Category,
			"count", len(r.products),
			"latency_ms", r.ms,
			"error", errString(r.err),
		)
		if r.err != nil {
			out.Errs = append(out.Errs, fmt.Errorf("%s: %w", r.name, r.err))
			continue
		}
		for i := range r.products {
			EnsureProductDefaults(&r.products[i])
		}
		out.Products = append(out.Products, r.products...)
	}
	return out
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
