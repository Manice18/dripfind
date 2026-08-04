package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/manice18/dripfind/backend/internal/models"
)

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type BewakoofProvider struct {
	session *sessionClient
}

func NewBewakoofProvider() *BewakoofProvider {
	return &BewakoofProvider{session: newSessionClient()}
}

func (p *BewakoofProvider) Name() string { return "bewakoof" }

func (p *BewakoofProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	query := item.SearchQuery
	if query == "" {
		query = BuildQuery(item, gender, "")
	}

	p.session.warm(ctx, "https://www.bewakoof.com/")

	searchURL := "https://www.bewakoof.com/search?q=" + url.QueryEscape(query)
	body, err := p.session.get(ctx, searchURL, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if err != nil {
		return nil, fmt.Errorf("bewakoof fetch: %w", err)
	}

	m := nextDataRe.FindSubmatch(body)
	if len(m) < 2 {
		return nil, fmt.Errorf("bewakoof: missing embedded product data")
	}

	var parsed struct {
		Props struct {
			PageProps struct {
				Data struct {
					Products []struct {
						Name         string  `json:"name"`
						CustomName   string  `json:"custom_name"`
						Brand        string  `json:"brand"`
						BrandName    string  `json:"brand_name"`
						Price        float64 `json:"price"`
						SellingPrice float64 `json:"sellingPrice"`
						ProductImage string  `json:"productImage"`
						ProductURL   string  `json:"productUrl"`
						URL          string  `json:"url"`
					} `json:"products"`
				} `json:"data"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal(m[1], &parsed); err != nil {
		return nil, fmt.Errorf("bewakoof decode: %w", err)
	}

	out := make([]models.Product, 0, 8)
	for _, r := range parsed.Props.PageProps.Data.Products {
		if len(out) >= 8 {
			break
		}
		title := strings.TrimSpace(r.Name)
		if title == "" {
			title = strings.TrimSpace(r.CustomName)
		}
		if title == "" {
			continue
		}
		price := r.Price
		if price == 0 {
			price = r.SellingPrice
		}
		brand := r.BrandName
		if brand == "" {
			brand = r.Brand
		}
		slug := r.ProductURL
		if slug == "" {
			slug = r.URL
		}
		link := bewakoofProductURL(slug)
		if link == "" {
			link = searchURL
		}

		out = append(out, models.Product{
			Title:    title,
			Brand:    nonempty(brand, "Bewakoof"),
			Price:    price,
			Currency: "INR",
			Website:  "bewakoof",
			URL:      link,
			Image:    r.ProductImage,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("bewakoof: no products for %q", query)
	}
	return out, nil
}

// bewakoofProductURL normalizes PDP paths. Bewakoof product pages are under /p/<slug>;
// search payloads often return just the slug (or a relative path without /p/), which
// 404s if joined as https://www.bewakoof.com/<slug>.
func bewakoofProductURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Already absolute.
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return raw
		}
		u.Path = ensureBewakoofPDPPath(u.Path)
		return u.String()
	}

	path := ensureBewakoofPDPPath(raw)
	return "https://www.bewakoof.com" + path
}

func ensureBewakoofPDPPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Already a product page (or another non-root path we should keep).
	lower := strings.ToLower(path)
	if strings.HasPrefix(lower, "/p/") || strings.HasPrefix(lower, "/search") {
		return path
	}
	// Bare slug from API → /p/<slug>
	return "/p" + path
}
