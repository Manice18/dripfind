package search

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/manice18/outfit_finder/backend/internal/models"
)

type AmazonProvider struct {
	session *sessionClient
}

func NewAmazonProvider() *AmazonProvider {
	return &AmazonProvider{session: newSessionClient()}
}

func (p *AmazonProvider) Name() string { return "amazon" }

func (p *AmazonProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	query := item.SearchQuery
	if query == "" {
		query = BuildQuery(item, gender, "")
	}

	p.session.warm(ctx, "https://www.amazon.in/")

	searchURL := "https://www.amazon.in/s?k=" + url.QueryEscape(query)
	body, err := p.session.get(ctx, searchURL, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if err != nil {
		return nil, fmt.Errorf("amazon fetch: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	out := make([]models.Product, 0, 8)
	doc.Find(`div[data-component-type="s-search-result"]`).Each(func(_ int, s *goquery.Selection) {
		if len(out) >= 8 {
			return
		}
		asin, _ := s.Attr("data-asin")
		if strings.TrimSpace(asin) == "" {
			return
		}

		title := strings.TrimSpace(s.Find("h2 a span").First().Text())
		if title == "" {
			title = strings.TrimSpace(s.Find("h2 span").First().Text())
		}
		if title == "" {
			return
		}

		href, _ := s.Find("h2 a").First().Attr("href")
		if href == "" {
			href, _ = s.Find("a.a-link-normal").First().Attr("href")
		}
		productURL := absolutize("https://www.amazon.in", href)
		if productURL == "" {
			productURL = searchURL
		}

		img, _ := s.Find("img.s-image").First().Attr("src")
		if img == "" {
			img, _ = s.Find("img").First().Attr("src")
		}

		priceText := strings.TrimSpace(s.Find(".a-price .a-offscreen").First().Text())
		if priceText == "" {
			whole := strings.TrimSpace(s.Find(".a-price-whole").First().Text())
			frac := strings.TrimSpace(s.Find(".a-price-fraction").First().Text())
			priceText = whole + frac
		}

		out = append(out, models.Product{
			Title:    title,
			Brand:    brandFromTitle(title),
			Price:    parseINRPrice(priceText),
			Currency: "INR",
			Website:  "amazon",
			URL:      productURL,
			Image:    img,
		})
	})

	if len(out) == 0 {
		return nil, fmt.Errorf("amazon: no products parsed for %q (often blocked; try SERPAPI_KEY)", query)
	}
	return out, nil
}

func absolutize(base, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	u, err := url.Parse(base)
	if err != nil {
		return base + href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return base + href
	}
	return u.ResolveReference(ref).String()
}
