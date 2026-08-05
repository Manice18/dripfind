package search

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// providerPacer enforces a minimum gap between requests to the same retailer
// (and optional shared CDN groups like Shopify), with jitter so timing is not
// perfectly metronomic.
type providerPacer struct {
	mu       sync.Mutex
	last     map[string]time.Time
	interval map[string]time.Duration
	fallback time.Duration
}

func newProviderPacer(fallback time.Duration, overrides map[string]time.Duration) *providerPacer {
	if fallback <= 0 {
		fallback = 400 * time.Millisecond
	}
	cp := make(map[string]time.Duration, len(overrides))
	for k, v := range overrides {
		cp[k] = v
	}
	return &providerPacer{
		last:     make(map[string]time.Time),
		interval: cp,
		fallback: fallback,
	}
}

func defaultProviderIntervals() map[string]time.Duration {
	return map[string]time.Duration{
		// Shopify storefronts — longer gaps; they share CDN reputation per IP.
		"snitch":      500 * time.Millisecond,
		"veirdo":      750 * time.Millisecond,
		"rarerabbit":  1 * time.Second,
		"westside":    700 * time.Millisecond,
		"offduty":     650 * time.Millisecond,
		"freakins":    650 * time.Millisecond,
		"pantproject": 650 * time.Millisecond,
		"bearhouse":   650 * time.Millisecond,
		"powerlook":   650 * time.Millisecond,
		"bluorng":     700 * time.Millisecond,
		// Shared bucket across all Shopify brands (most important for 429s).
		"group:shopify": 550 * time.Millisecond,
		// Big marketplaces — a bit gentler than before.
		"myntra":   400 * time.Millisecond,
		"ajio":     350 * time.Millisecond,
		"flipkart": 500 * time.Millisecond,
		"hm":       800 * time.Millisecond,
		"bewakoof": 350 * time.Millisecond,
	}
}

func paceKeys(provider string) []string {
	keys := []string{provider}
	if isShopifyBrand(provider) {
		keys = append(keys, "group:shopify")
	}
	return keys
}

func isShopifyBrand(name string) bool {
	switch name {
	case "snitch", "veirdo", "westside", "offduty", "freakins",
		"pantproject", "bearhouse", "powerlook", "bluorng", "rarerabbit":
		return true
	default:
		return false
	}
}

func (p *providerPacer) intervalFor(key string) time.Duration {
	if d, ok := p.interval[key]; ok && d > 0 {
		return d
	}
	return p.fallback
}

// jitterDuration returns base with ±40% randomness (never below 40% of base).
func jitterDuration(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	// factor in [0.6, 1.4]
	factor := 0.6 + rand.Float64()*0.8
	d := time.Duration(float64(base) * factor)
	if d < base/5 {
		d = base / 5
	}
	return d
}

// Wait blocks until every pace key for this provider is due.
func (p *providerPacer) Wait(ctx context.Context, provider string) error {
	for _, key := range paceKeys(provider) {
		if err := p.waitKey(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *providerPacer) waitKey(ctx context.Context, key string) error {
	for {
		p.mu.Lock()
		interval := jitterDuration(p.intervalFor(key))
		last := p.last[key]
		wait := interval - time.Since(last)
		if last.IsZero() || wait <= 0 {
			p.last[key] = time.Now()
			p.mu.Unlock()
			return nil
		}
		p.mu.Unlock()

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
			// loop to re-check under lock (another goroutine may have claimed the slot)
		}
	}
}
