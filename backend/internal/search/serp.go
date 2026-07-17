package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

// SerpShoppingProvider uses SerpAPI Google Shopping when SERPAPI_KEY is set.
type SerpShoppingProvider struct {
	apiKey string
	client *http.Client
}

func NewSerpShoppingProvider(apiKey string) *SerpShoppingProvider {
	return &SerpShoppingProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *SerpShoppingProvider) Name() string { return "google_shopping" }

func (p *SerpShoppingProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	query := item.SearchQuery
	if query == "" {
		query = BuildQuery(item, gender, "")
	}

	u := url.URL{
		Scheme: "https",
		Host:   "serpapi.com",
		Path:   "/search.json",
	}
	q := u.Query()
	q.Set("engine", "google_shopping")
	q.Set("q", query)
	q.Set("api_key", p.apiKey)
	q.Set("gl", "in")
	q.Set("hl", "en")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("serpapi status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var parsed struct {
		ShoppingResults []struct {
			Title      string  `json:"title"`
			Source     string  `json:"source"`
			Price      string  `json:"price"`
			Extracted  float64 `json:"extracted_price"`
			Link       string  `json:"link"`
			ProductLink string `json:"product_link"`
			Thumbnail  string  `json:"thumbnail"`
		} `json:"shopping_results"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	out := make([]models.Product, 0, 8)
	for i, r := range parsed.ShoppingResults {
		if i >= 8 {
			break
		}
		link := r.ProductLink
		if link == "" {
			link = r.Link
		}
		out = append(out, models.Product{
			Title:    r.Title,
			Brand:    r.Source,
			Price:    r.Extracted,
			Currency: "INR",
			Website:  normalizeSource(r.Source),
			URL:      link,
			Image:    r.Thumbnail,
		})
	}
	return out, nil
}

func normalizeSource(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "amazon"):
		return "amazon"
	case strings.Contains(s, "myntra"):
		return "myntra"
	case strings.Contains(s, "ajio"):
		return "ajio"
	case strings.Contains(s, "flipkart"):
		return "flipkart"
	default:
		return "google_shopping"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
