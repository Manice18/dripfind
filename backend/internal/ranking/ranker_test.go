package ranking

import (
	"strings"
	"testing"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

func TestRejectWomensWhenMale(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Shirt", Color: "White", Pattern: "Striped", Fit: "Regular", Material: "Cotton"}
	products := []models.Product{
		{Title: "Men's White & Grey Striped Cotton Shirt", Brand: "Bewakoof", Website: "bewakoof", Price: 1599, Image: "a"},
		{Title: "Rareism Women's Cuba Light Beige Cotton Shirt", Brand: "Rareism", Website: "rarerabbit", Price: 1249, Image: "b"},
		{Title: "WROGN Tie and Dye Slim Fit Cotton Casual Shirt", Brand: "WROGN", Website: "myntra", Price: 999, Image: "c"},
	}
	out := r.Rank(item, products, "male")
	if len(out) != 1 {
		t.Fatalf("expected 1 product, got %d %#v", len(out), out)
	}
	if out[0].Brand != "Bewakoof" {
		t.Fatalf("expected Bewakoof, got %s", out[0].Brand)
	}
}

func TestRejectChinoWhenDenim(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Trousers", Color: "Black", Material: "Denim", Fit: "Relaxed", Pattern: "Solid"}
	products := []models.Product{
		{Title: "Snitch Black Baggy Jeans", Brand: "Snitch", Website: "snitch", Price: 2039, Image: "a"},
		{Title: "Bearhouse Black Solid Slim Fit Chino Trousers", Brand: "Bearhouse", Website: "bearhouse", Price: 2799, Image: "b"},
		{Title: "WES Formals Black Relaxed-Fit Mid-Rise Formal Trousers", Brand: "Westside", Website: "westside", Price: 999, Image: "c"},
	}
	out := r.Rank(item, products, "male")
	if len(out) != 1 || out[0].Brand != "Snitch" {
		t.Fatalf("expected only Snitch baggy jeans, got %#v", out)
	}
}

func TestRejectBabyAndTShirt(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Shirt", Color: "White", Pattern: "Striped", Fit: "Regular", Material: "Cotton"}
	products := []models.Product{
		{Title: "Men's White Striped Cotton Linen Shirt", Brand: "Bewakoof", Website: "bewakoof", Price: 1499, Image: "a"},
		{Title: "HOP Baby Boys White Striped Cotton Shirt", Brand: "HOP BABY", Website: "westside", Price: 599, Image: "b"},
		{Title: "White Striped Cotton-Blend T-Shirt", Brand: "NUOFLEXX", Website: "westside", Price: 799, Image: "c"},
		{Title: "Snitch White Striped Casual Shirt", Brand: "Snitch", Website: "snitch", Price: 999, Image: "d"},
	}
	out := r.Rank(item, products, "male")
	if len(out) != 2 {
		t.Fatalf("expected 2 adult shirts, got %d %#v", len(out), out)
	}
	for _, p := range out {
		if strings.Contains(strings.ToLower(p.Title), "baby") || strings.Contains(strings.ToLower(p.Title), "t-shirt") {
			t.Fatalf("unexpected product: %s", p.Title)
		}
	}
}

func TestDiversifySites(t *testing.T) {
	r := NewScoreRanker()
	r.MaxPerWebsite = 2
	r.Limit = 6
	item := models.ClothingItem{Category: "Shirt", Color: "White", Pattern: "Striped", Fit: "Regular", Material: "Cotton"}
	products := []models.Product{}
	for i := 0; i < 5; i++ {
		products = append(products, models.Product{
			Title: "Men White Striped Casual Shirt " + string(rune('A'+i)), Brand: "Bewakoof", Website: "bewakoof", Price: 1000, Image: "b" + string(rune('0'+i)),
		})
	}
	products = append(products,
		models.Product{Title: "Men White Striped Casual Shirt X", Brand: "Snitch", Website: "snitch", Price: 1100, Image: "s1"},
		models.Product{Title: "Men White Striped Casual Shirt Y", Brand: "Powerlook", Website: "powerlook", Price: 1200, Image: "p1"},
		models.Product{Title: "Men White Striped Casual Shirt Z", Brand: "Ajio", Website: "ajio", Price: 1300, Image: "a1"},
	)
	out := r.Rank(item, products, "male")
	counts := map[string]int{}
	for _, p := range out {
		counts[p.Website]++
	}
	if counts["bewakoof"] > 2 {
		t.Fatalf("bewakoof dominated: %#v", counts)
	}
	if len(counts) < 3 {
		t.Fatalf("expected multiple sites, got %#v", counts)
	}
}
