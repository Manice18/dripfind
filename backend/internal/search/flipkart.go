package search

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/manice18/outfit_finder/backend/internal/models"
)

var flipkartPriceRe = regexp.MustCompile(`₹\s*[\d,]+`)

type FlipkartProvider struct {
	session *sessionClient
}

func NewFlipkartProvider() *FlipkartProvider {
	return &FlipkartProvider{session: newSessionClient()}
}

func (p *FlipkartProvider) Name() string { return "flipkart" }

func (p *FlipkartProvider) Search(ctx context.Context, item models.ClothingItem, gender string) ([]models.Product, error) {
	query := item.SearchQuery
	if query == "" {
		query = BuildQuery(item, gender)
	}

	p.session.warm(ctx, "https://www.flipkart.com/")

	searchURL := "https://www.flipkart.com/search?q=" + url.QueryEscape(query)
	body, err := p.session.get(ctx, searchURL, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if err != nil {
		return nil, fmt.Errorf("flipkart fetch: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	out := make([]models.Product, 0, 8)
	seen := map[string]struct{}{}

	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if len(out) >= 8 {
			return
		}
		href, ok := s.Attr("href")
		if !ok || !strings.Contains(href, "/p/") {
			return
		}
		productURL := absolutize("https://www.flipkart.com", href)
		if _, exists := seen[productURL]; exists {
			return
		}

		img, _ := s.Find("img").First().Attr("src")
		if img == "" {
			img, _ = s.Find("img").First().Attr("data-src")
		}
		if img == "" || (!strings.Contains(img, "rukminim") && !strings.Contains(img, "flixcart")) {
			return
		}

		title, _ := s.Find("img").First().Attr("alt")
		title = strings.TrimSpace(title)
		if title == "" {
			title = strings.TrimSpace(s.Text())
		}
		title = strings.Join(strings.Fields(title), " ")
		if len(title) < 8 {
			return
		}

		priceText := flipkartPriceRe.FindString(s.Text())
		if priceText == "" {
			// look at parent card
			priceText = flipkartPriceRe.FindString(s.Parent().Parent().Text())
		}

		seen[productURL] = struct{}{}
		out = append(out, models.Product{
			Title:    title,
			Brand:    brandFromTitle(title),
			Price:    parseINRPrice(priceText),
			Currency: "INR",
			Website:  "flipkart",
			URL:      productURL,
			Image:    img,
		})
	})

	if len(out) == 0 {
		return nil, fmt.Errorf("flipkart: no products for %q", query)
	}
	return out, nil
}
