package search

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

type sessionClient struct {
	client *http.Client
	warmed map[string]bool
	mu     sync.Mutex
}

func newSessionClient() *sessionClient {
	jar, _ := cookiejar.New(nil)
	return &sessionClient{
		client: &http.Client{
			Timeout:   25 * time.Second,
			Jar:       jar,
			Transport: &dynamicTransport{},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 8 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		warmed: map[string]bool{},
	}
}

// dynamicTransport picks up ConfigureScrape proxy changes without recreating clients.
type dynamicTransport struct{}

func (d *dynamicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return activeTransport().RoundTrip(req)
}

func (s *sessionClient) warm(ctx context.Context, homeURL string) {
	s.mu.Lock()
	if s.warmed[homeURL] {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, homeURL, nil)
	if err != nil {
		return
	}
	setBrowserHeaders(req, "text/html")
	if err := activeGate().acquire(ctx); err != nil {
		return
	}
	resp, err := s.client.Do(req)
	activeGate().release()
	if err != nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if resp.StatusCode < 400 {
		s.mu.Lock()
		s.warmed[homeURL] = true
		s.mu.Unlock()
	}
}

func (s *sessionClient) get(ctx context.Context, rawURL, accept string) ([]byte, error) {
	return doGatedGet(ctx, s.client, rawURL, accept, nil)
}

func newHTTPClient() *http.Client {
	return newSessionClient().client
}

func fetch(ctx context.Context, client *http.Client, rawURL string, headers map[string]string) ([]byte, error) {
	return doGatedGet(ctx, client, rawURL, "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8", headers)
}

func doGatedGet(ctx context.Context, client *http.Client, rawURL, accept string, headers map[string]string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if err := activeGate().acquire(ctx); err != nil {
			return nil, err
		}
		body, status, err := rawGet(ctx, client, rawURL, accept, headers)
		activeGate().release()
		if err != nil {
			return nil, err
		}
		if status == 429 && attempt == 0 {
			lastErr = &httpStatusError{Code: status}
			delay := time.Duration(1000+rand.Intn(1000)) * time.Millisecond
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return nil, ctx.Err()
			case <-t.C:
			}
			continue
		}
		if status >= 400 {
			return nil, &httpStatusError{Code: status}
		}
		return body, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("fetch failed")
}

func rawGet(ctx context.Context, client *http.Client, rawURL, accept string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	setBrowserHeaders(req, accept)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

var userAgents = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
}

func setBrowserHeaders(req *http.Request, accept string) {
	ua := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept-Language", "en-IN,en;q=0.9")
	req.Header.Set("Accept", accept)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="131", "Not_A Brand";v="24", "Google Chrome";v="131"`)
	mobile := strings.Contains(ua, "iPhone") || strings.Contains(ua, "Mobile")
	if mobile {
		req.Header.Set("Sec-Ch-Ua-Mobile", "?1")
		req.Header.Set("Sec-Ch-Ua-Platform", `"iOS"`)
	} else {
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	}
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	if strings.Contains(accept, "json") {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
	} else {
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Sec-Fetch-User", "?1")
	}
}

func parseINRPrice(raw string) float64 {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", "")
	raw = strings.ReplaceAll(raw, "₹", "")
	raw = strings.ReplaceAll(raw, "Rs.", "")
	raw = strings.ReplaceAll(raw, "Rs", "")
	raw = strings.TrimSpace(raw)
	var n float64
	var dec float64
	var inDec bool
	var places float64 = 1
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			if !inDec {
				n = n*10 + float64(r-'0')
			} else if places <= 100 {
				dec = dec*10 + float64(r-'0')
				places *= 10
			}
		} else if r == '.' && !inDec {
			inDec = true
		} else if n > 0 {
			break
		}
	}
	if inDec && places > 1 {
		n += dec / (places / 10)
	}
	return n
}

func brandFromTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Unknown"
	}
	parts := strings.Fields(title)
	if len(parts) == 0 {
		return "Unknown"
	}
	first := parts[0]
	if len(parts) >= 2 && (parts[1] == "-" || strings.EqualFold(parts[1], "by")) {
		return first
	}
	if len(first) <= 18 {
		return first
	}
	return first
}

func absolutize(base, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	return strings.TrimRight(base, "/") + href
}
