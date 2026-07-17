package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

type AjioProvider struct {
	session *sessionClient
}

func NewAjioProvider() *AjioProvider {
	return &AjioProvider{session: newSessionClient()}
}

func (p *AjioProvider) Name() string { return "ajio" }

func (p *AjioProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	query := item.SearchQuery
	if query == "" {
		query = BuildQuery(item, gender, "")
	}

	apiURL := "https://www.ajio.com/api/search?" + url.Values{
		"fields":      {"SITE"},
		"currentPage": {"0"},
		"pageSize":    {"20"},
		"format":      {"json"},
		"query":       {query + ":relevance"},
		"sortBy":      {"relevance"},
		"text":        {query},
	}.Encode()

	p.session.warm(ctx, "https://www.ajio.com/")
	body, err := p.session.get(ctx, apiURL, "application/json, text/plain, */*")
	if err != nil {
		return nil, fmt.Errorf("ajio fetch: %w", err)
	}

	var parsed struct {
		Products []struct {
			Name      string `json:"name"`
			BrandName string `json:"brandName"`
			Brand     string `json:"brand"`
			Price     struct {
				Value float64 `json:"value"`
			} `json:"price"`
			WasPriceData struct {
				Value float64 `json:"value"`
			} `json:"wasPriceData"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			URL  string `json:"url"`
			Code string `json:"code"`
		} `json:"products"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("ajio decode: %w", err)
	}

	out := make([]models.Product, 0, 8)
	for _, r := range parsed.Products {
		if len(out) >= 8 {
			break
		}
		title := strings.TrimSpace(r.Name)
		if title == "" {
			continue
		}
		brand := r.BrandName
		if brand == "" {
			brand = r.Brand
		}
		price := r.Price.Value
		if price == 0 {
			price = r.WasPriceData.Value
		}
		img := ""
		if len(r.Images) > 0 {
			img = r.Images[0].URL
		}
		link := r.URL
		if link != "" && !strings.HasPrefix(link, "http") {
			link = "https://www.ajio.com" + link
		}
		if link == "" {
			link = "https://www.ajio.com/search/?text=" + url.QueryEscape(query)
		}

		out = append(out, models.Product{
			Title:    title,
			Brand:    nonempty(brand, brandFromTitle(title)),
			Price:    price,
			Currency: "INR",
			Website:  "ajio",
			URL:      link,
			Image:    img,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("ajio: no products for %q", query)
	}
	return out, nil
}
