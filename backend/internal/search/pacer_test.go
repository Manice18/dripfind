package search

import (
	"context"
	"testing"
	"time"
)

func TestProviderPacerSpacesSameKey(t *testing.T) {
	p := newProviderPacer(50*time.Millisecond, map[string]time.Duration{
		"snitch": 80 * time.Millisecond,
	})
	ctx := context.Background()
	start := time.Now()
	if err := p.Wait(ctx, "snitch"); err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(ctx, "snitch"); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	// With jitter ≥ 60% of 80ms → at least ~48ms between claims.
	if elapsed < 40*time.Millisecond {
		t.Fatalf("expected pacing, got %s", elapsed)
	}
}

func TestProviderPacerShopifyGroupSerializesBrands(t *testing.T) {
	p := newProviderPacer(10*time.Millisecond, map[string]time.Duration{
		"snitch":        10 * time.Millisecond,
		"veirdo":        10 * time.Millisecond,
		"group:shopify": 60 * time.Millisecond,
	})
	ctx := context.Background()
	start := time.Now()
	if err := p.Wait(ctx, "snitch"); err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(ctx, "veirdo"); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 35*time.Millisecond {
		t.Fatalf("expected shopify group pacing between brands, got %s", elapsed)
	}
}

func TestJitterDurationVaries(t *testing.T) {
	base := 200 * time.Millisecond
	seen := map[time.Duration]bool{}
	for i := 0; i < 40; i++ {
		seen[jitterDuration(base)] = true
	}
	if len(seen) < 3 {
		t.Fatalf("expected varied jitter, got %d distinct values", len(seen))
	}
}
