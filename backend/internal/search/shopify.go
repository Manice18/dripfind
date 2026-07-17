package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

// ShopifyProvider searches any Shopify storefront via /search/suggest.json.
type ShopifyProvider struct {
	website string
	baseURL string
	session *sessionClient
}

func NewShopifyProvider(website, baseURL string) *ShopifyProvider {
	return &ShopifyProvider{
		website: website,
		baseURL: strings.TrimRight(baseURL, "/"),
		session: newSessionClient(),
	}
}

func (p *ShopifyProvider) Name() string { return p.website }

func (p *ShopifyProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	queries := shopifyQueries(item, gender)
	p.session.warm(ctx, p.baseURL+"/")

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
		lastErr = fmt.Errorf("%s: no products for %q", p.website, query)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("%s: no products", p.website)
}

func (p *ShopifyProvider) searchQuery(ctx context.Context, query string) ([]models.Product, error) {
	apiURL := fmt.Sprintf(
		"%s/search/suggest.json?q=%s&resources[type]=product&resources[limit]=10",
		p.baseURL,
		url.QueryEscape(query),
	)

	body, err := p.session.get(ctx, apiURL, "application/json")
	if err != nil {
		return nil, fmt.Errorf("%s fetch: %w", p.website, err)
	}

	var parsed struct {
		Resources struct {
			Results struct {
				Products []shopifySuggestProduct `json:"products"`
			} `json:"results"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("%s decode: %w", p.website, err)
	}

	out := make([]models.Product, 0, 8)
	for _, r := range parsed.Resources.Results.Products {
		if len(out) >= 8 {
			break
		}
		title := strings.TrimSpace(r.Title)
		if title == "" {
			continue
		}
		img := r.Image
		if img == "" && r.FeaturedImage != nil {
			img = r.FeaturedImage.URL
		}
		link := r.URL
		if link != "" && !strings.HasPrefix(link, "http") {
			link = p.baseURL + link
		}
		if link == "" && r.Handle != "" {
			link = p.baseURL + "/products/" + r.Handle
		}
		if link == "" {
			continue
		}

		out = append(out, models.Product{
			Title:    title,
			Brand:    nonempty(r.Vendor, brandFromTitle(title)),
			Price:    parseShopifyPrice(r.Price),
			Currency: "INR",
			Website:  p.website,
			URL:      link,
			Image:    img,
		})
	}
	return out, nil
}

// shopifyQueries returns progressively shorter search strings.
// Brand Shopify catalogs often miss long attribute-heavy queries.
func shopifyQueries(item models.ClothingItem, gender string) []string {
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
		if _, ok := seen[strings.ToLower(q)]; ok {
			continue
		}
		seen[strings.ToLower(q)] = struct{}{}
		out = append(out, q)
	}
	return out
}

type shopifySuggestProduct struct {
	Title         string  `json:"title"`
	Vendor        string  `json:"vendor"`
	Handle        string  `json:"handle"`
	URL           string  `json:"url"`
	Image         string  `json:"image"`
	Price         any     `json:"price"` // string or number depending on theme
	FeaturedImage *struct {
		URL string `json:"url"`
	} `json:"featured_image"`
}

func parseShopifyPrice(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		return parseINRPrice(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	default:
		s := fmt.Sprint(v)
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return parseINRPrice(s)
	}
}

// HomegrownShopifyBrands are Indian fashion stores from the curated list
// that expose Shopify predictive search.
func HomegrownShopifyBrands() []Provider {
	return []Provider{
		NewShopifyProvider("snitch", "https://www.snitch.co.in"),
		NewShopifyProvider("veirdo", "https://veirdo.in"),
		NewShopifyProvider("westside", "https://www.westside.com"),
		NewShopifyProvider("offduty", "https://offduty.in"),
		NewShopifyProvider("freakins", "https://freakins.com"),
		NewShopifyProvider("pantproject", "https://thepantproject.com"),
		NewShopifyProvider("bearhouse", "https://www.thebearhouse.com"),
		NewShopifyProvider("powerlook", "https://powerlook.in"),
		NewShopifyProvider("bluorng", "https://bluorng.com"),
		NewShopifyProvider("rarerabbit", "https://rareism.com"),
	}
}
