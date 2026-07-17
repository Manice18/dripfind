package search

import (
	"context"
	"fmt"
	"io"
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
			Timeout: 25 * time.Second,
			Jar:     jar,
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

func (s *sessionClient) warm(ctx context.Context, homeURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.warmed[homeURL] {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, homeURL, nil)
	if err != nil {
		return
	}
	setBrowserHeaders(req, "text/html")
	resp, err := s.client.Do(req)
	if err != nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	s.warmed[homeURL] = true
}

func (s *sessionClient) get(ctx context.Context, rawURL, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	setBrowserHeaders(req, accept)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return body, nil
}

func setBrowserHeaders(req *http.Request, accept string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-IN,en;q=0.9")
	req.Header.Set("Accept", accept)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="131", "Not_A Brand";v="24", "Google Chrome";v="131"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
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

func newHTTPClient() *http.Client {
	return newSessionClient().client
}

func fetch(ctx context.Context, client *http.Client, rawURL string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	setBrowserHeaders(req, "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return body, nil
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
