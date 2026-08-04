package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/manice18/dripfind/backend/internal/models"
)

const hmBase = "https://www2.hm.com"

// HMProvider searches H&M India via the search-results page __NEXT_DATA__ payload.
type HMProvider struct {
	session *sessionClient
}

func NewHMProvider() *HMProvider {
	return &HMProvider{session: newSessionClient()}
}

func (p *HMProvider) Name() string { return "hm" }

func (p *HMProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	queries := hmQueries(item, gender)
	p.session.warm(ctx, hmBase+"/en_in/")

	var lastErr error
	for _, query := range queries {
		products, err := p.searchQuery(ctx, query)
		if err != nil {
			lastErr = err
			continue
		}
		if len(products) > 0 {
			return products, nil
		}
		lastErr = fmt.Errorf("hm: no products for %q", query)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("hm: no products")
}

func (p *HMProvider) searchQuery(ctx context.Context, query string) ([]models.Product, error) {
	searchURL := fmt.Sprintf(
		"%s/en_in/search-results.html?q=%s",
		hmBase,
		url.QueryEscape(query),
	)

	body, err := p.session.get(ctx, searchURL, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if err != nil {
		return nil, fmt.Errorf("hm fetch: %w", err)
	}
	if strings.Contains(string(body), "Access Denied") {
		return nil, fmt.Errorf("hm fetch: access denied")
	}

	m := nextDataRe.FindSubmatch(body)
	if len(m) < 2 {
		return nil, fmt.Errorf("hm: missing embedded product data")
	}

	var parsed struct {
		Props struct {
			PageProps struct {
				SrpProps struct {
					Hits []hmHit `json:"hits"`
				} `json:"srpProps"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal(m[1], &parsed); err != nil {
		return nil, fmt.Errorf("hm decode: %w", err)
	}

	out := make([]models.Product, 0, 8)
	for _, h := range parsed.Props.PageProps.SrpProps.Hits {
		if len(out) >= 8 {
			break
		}
		title := strings.TrimSpace(h.Title)
		if title == "" {
			continue
		}
		link := strings.TrimSpace(h.PdpURL)
		if link == "" && h.ArticleCode != "" {
			link = fmt.Sprintf("%s/en_in/productpage.%s.html", hmBase, h.ArticleCode)
		}
		if link == "" {
			continue
		}
		if strings.HasPrefix(link, "/") {
			link = hmBase + link
		}

		img := h.ImageProductSrc
		if img == "" {
			img = h.ImageModelSrc
		}

		out = append(out, models.Product{
			Title:    title,
			Brand:    nonempty(h.BrandName, "H&M"),
			Price:    pickHMPrice(h.Prices),
			Currency: "INR",
			Website:  "hm",
			URL:      link,
			Image:    img,
		})
	}
	return out, nil
}

type hmHit struct {
	Title           string       `json:"title"`
	BrandName       string       `json:"brandName"`
	PdpURL          string       `json:"pdpUrl"`
	ArticleCode     string       `json:"articleCode"`
	ImageProductSrc string       `json:"imageProductSrc"`
	ImageModelSrc   string       `json:"imageModelSrc"`
	Prices          []hmPriceRow `json:"prices"`
}

type hmPriceRow struct {
	Price     float64 `json:"price"`
	PriceType string  `json:"priceType"`
}

func pickHMPrice(prices []hmPriceRow) float64 {
	var white, red float64
	for _, p := range prices {
		switch p.PriceType {
		case "redPrice":
			red = p.Price
		case "whitePrice":
			white = p.Price
		default:
			if white == 0 {
				white = p.Price
			}
		}
	}
	if red > 0 {
		return red
	}
	return white
}

func hmQueries(item models.ClothingItem, gender string) []string {
	full := item.SearchQuery
	if full == "" {
		full = BuildQuery(item, gender, "")
	}

	parts := []string{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || strings.EqualFold(s, "solid") || strings.EqualFold(s, "other") {
			return
		}
		parts = append(parts, s)
	}
	switch strings.ToLower(strings.TrimSpace(gender)) {
	case "male", "men", "man":
		add("men")
	case "female", "women", "woman":
		add("women")
	}
	add(item.Color)
	if !strings.EqualFold(item.Pattern, "Solid") {
		add(item.Pattern)
	}
	add(item.Category)

	short := strings.Join(parts, " ")
	shortest := strings.TrimSpace(item.Category)
	if shortest == "" {
		shortest = "clothing"
	}

	out := []string{}
	seen := map[string]struct{}{}
	for _, q := range []string{full, short, shortest} {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		key := strings.ToLower(q)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, q)
	}
	return out
}
