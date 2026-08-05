package search

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/manice18/dripfind/backend/internal/models"
)

type Provider interface {
	Name() string
	Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error)
}

// providerTimeout caps how long any single store can block an item search.
// Includes per-provider pacing + scrape-gate wait.
const providerTimeout = 25 * time.Second

type Engine struct {
	providers []Provider
	log       *slog.Logger
	breaker   *circuitBreaker
	pacer     *providerPacer
}

func NewEngine(log *slog.Logger, providers ...Provider) *Engine {
	return &Engine{
		providers: providers,
		log:       log,
		breaker:   newCircuitBreaker(2, 2*time.Minute),
		pacer:     newProviderPacer(400*time.Millisecond, defaultProviderIntervals()),
	}
}

// SetBreakerConfig overrides circuit-breaker defaults (threshold trips, cooldown).
func (e *Engine) SetBreakerConfig(threshold int, cooldown time.Duration) {
	e.breaker = newCircuitBreaker(threshold, cooldown)
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
		skipped  bool
	}

	ch := make(chan providerOut, len(e.providers))
	var wg sync.WaitGroup

	for _, p := range e.providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			name := prov.Name()
			if !e.breaker.Allow(name) {
				ch <- providerOut{name: name, err: ErrCircuitOpen, skipped: true}
				return
			}
			start := time.Now()
			pctx, cancel := context.WithTimeout(ctx, providerTimeout)
			defer cancel()

			if err := e.pacer.Wait(pctx, name); err != nil {
				ch <- providerOut{name: name, err: err, ms: time.Since(start).Milliseconds()}
				return
			}

			products, err := prov.Search(pctx, item, gender)
			if err != nil {
				e.breaker.Failure(name, err)
			} else {
				e.breaker.Success(name)
			}
			ch <- providerOut{
				name:     name,
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
			"skipped", r.skipped,
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
