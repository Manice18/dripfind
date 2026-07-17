package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

type MyntraProvider struct {
	session *sessionClient
}

func NewMyntraProvider() *MyntraProvider {
	return &MyntraProvider{session: newSessionClient()}
}

func (p *MyntraProvider) Name() string { return "myntra" }

func (p *MyntraProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	query := item.SearchQuery
	if query == "" {
		query = BuildQuery(item, gender)
	}

	p.session.warm(ctx, "https://www.myntra.com/")

	slug := strings.ToLower(strings.Join(strings.Fields(query), "-"))
	apiURL := fmt.Sprintf(
		"https://www.myntra.com/gateway/v2/search/%s?rows=24&o=0&plaEnabled=false&rawQuery=%s",
		url.PathEscape(slug),
		url.QueryEscape(query),
	)

	body, err := p.session.get(ctx, apiURL, "application/json")
	if err != nil {
		return nil, fmt.Errorf("myntra fetch: %w", err)
	}

	var parsed struct {
		Products []struct {
			ProductName    string  `json:"productName"`
			Brand          string  `json:"brand"`
			SearchImage    string  `json:"searchImage"`
			Price          float64 `json:"price"`
			MRP            float64 `json:"mrp"`
			LandingPageURL string  `json:"landingPageUrl"`
			ProductID      int64   `json:"productId"`
		} `json:"products"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("myntra decode: %w", err)
	}

	out := make([]models.Product, 0, 8)
	for _, r := range parsed.Products {
		if len(out) >= 8 {
			break
		}
		title := strings.TrimSpace(r.Brand + " " + r.ProductName)
		if title == "" {
			continue
		}
		price := r.Price
		if price == 0 {
			price = r.MRP
		}
		link := r.LandingPageURL
		if link == "" && r.ProductID > 0 {
			link = fmt.Sprintf("https://www.myntra.com/%d", r.ProductID)
		}
		if link != "" && !strings.HasPrefix(link, "http") {
			link = "https://www.myntra.com/" + strings.TrimPrefix(link, "/")
		}
		if link == "" {
			link = "https://www.myntra.com/" + slug + "?rawQuery=" + url.QueryEscape(query)
		}

		out = append(out, models.Product{
			Title:    title,
			Brand:    nonempty(r.Brand, brandFromTitle(title)),
			Price:    price,
			Currency: "INR",
			Website:  "myntra",
			URL:      link,
			Image:    r.SearchImage,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("myntra: no products for %q", query)
	}
	return out, nil
}

func nonempty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
