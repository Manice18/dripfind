package search

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// scrapeGate limits how hard we hit retailer CDNs from one IP.
// VPS/datacenter IPs get 429 when ~15 providers fire at once; phones don't.
type scrapeGate struct {
	sem        chan struct{}
	minSpacing time.Duration
	mu         sync.Mutex
	lastStart  time.Time
}

func newScrapeGate(maxConcurrent int, minSpacing time.Duration) *scrapeGate {
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}
	if minSpacing < 0 {
		minSpacing = 0
	}
	return &scrapeGate{
		sem:        make(chan struct{}, maxConcurrent),
		minSpacing: minSpacing,
	}
}

func (g *scrapeGate) acquire(ctx context.Context) error {
	select {
	case g.sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	for {
		g.mu.Lock()
		spacing := jitterDuration(g.minSpacing)
		wait := spacing - time.Since(g.lastStart)
		if g.lastStart.IsZero() || wait <= 0 {
			g.lastStart = time.Now()
			g.mu.Unlock()
			return nil
		}
		g.mu.Unlock()

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			<-g.sem
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (g *scrapeGate) release() {
	select {
	case <-g.sem:
	default:
	}
}

var (
	gateMu sync.RWMutex
	gate   = newScrapeGate(2, 300*time.Millisecond)
)

// ConfigureScrape sets global outbound scrape pacing and optional HTTP(S) proxy.
// proxyURL example: http://user:pass@residential-proxy:8000
func ConfigureScrape(maxConcurrent int, minInterval time.Duration, proxyURL string) error {
	gateMu.Lock()
	gate = newScrapeGate(maxConcurrent, minInterval)
	gateMu.Unlock()

	if proxyURL == "" {
		setProxyTransport(nil)
		return nil
	}
	u, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("SCRAPE_PROXY_URL: %w", err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(u),
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	setProxyTransport(transport)
	return nil
}

func activeGate() *scrapeGate {
	gateMu.RLock()
	defer gateMu.RUnlock()
	return gate
}

var (
	proxyMu        sync.RWMutex
	proxyTransport http.RoundTripper
)

func setProxyTransport(t http.RoundTripper) {
	proxyMu.Lock()
	proxyTransport = t
	proxyMu.Unlock()
}

func activeTransport() http.RoundTripper {
	proxyMu.RLock()
	defer proxyMu.RUnlock()
	if proxyTransport != nil {
		return proxyTransport
	}
	return http.DefaultTransport
}
