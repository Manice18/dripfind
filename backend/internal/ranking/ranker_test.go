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
	out := r.Rank(item, products, "male", "Casual")
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
		{Title: "Snitch Men Black Baggy Jeans", Brand: "Snitch", Website: "snitch", Price: 2039, Image: "a"},
		{Title: "Bearhouse Black Solid Slim Fit Chino Trousers", Brand: "Bearhouse", Website: "bearhouse", Price: 2799, Image: "b"},
		{Title: "WES Formals Black Relaxed-Fit Mid-Rise Formal Trousers", Brand: "Westside", Website: "westside", Price: 999, Image: "c"},
	}
	out := r.Rank(item, products, "male", "Casual")
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
		{Title: "Men Snitch White Striped Casual Shirt", Brand: "Snitch", Website: "snitch", Price: 999, Image: "d"},
	}
	out := r.Rank(item, products, "male", "Casual")
	if len(out) != 2 {
		t.Fatalf("expected 2 adult shirts, got %d %#v", len(out), out)
	}
	for _, p := range out {
		if strings.Contains(strings.ToLower(p.Title), "baby") || strings.Contains(strings.ToLower(p.Title), "t-shirt") {
			t.Fatalf("unexpected product: %s", p.Title)
		}
	}
}

func TestRejectBlueTeeForWhiteTank(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Tank Top", Color: "White", Pattern: "Solid", Fit: "Regular", Material: "Cotton"}
	products := []models.Product{
		{Title: "Van Heusen Men Blue Round Neck Innerwear T-shirt", Brand: "Van Heusen", Website: "myntra", Price: 505, Image: "a"},
		{Title: "Men White Slim Fit Cotton Tank Top", Brand: "Decathlon", Website: "myntra", Price: 699, Image: "b"},
		{Title: "Men White Sleeveless Ribbed Tank", Brand: "Snitch", Website: "snitch", Price: 799, Image: "c"},
	}
	out := r.Rank(item, products, "male", "Casual")
	for _, p := range out {
		low := strings.ToLower(p.Title)
		if strings.Contains(low, "blue") || strings.Contains(low, "t-shirt") {
			t.Fatalf("unexpected product: %s", p.Title)
		}
	}
	if len(out) == 0 {
		t.Fatal("expected at least one white tank")
	}
}

func TestRejectWomensTankWithoutMenLabel(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Tank Top", Color: "White", Pattern: "Solid", Fit: "Slim", Material: "Cotton"}
	products := []models.Product{
		{Title: "Decathlon Domyos Men White Training Tank", Brand: "Decathlon", Website: "myntra", Price: 699, Image: "a"},
		{Title: "Superstar Women White Solid Cotton-Blend Tank Top", Brand: "SUPERSTAR", Website: "westside", Price: 399, Image: "b"},
		{Title: "White Slim Fit Corset Tank Top", Brand: "Off Duty", Website: "offduty", Price: 650, Image: "c"},
		{Title: "Men White Crystal Tank Top", Brand: "Bluorng", Website: "bluorng", Price: 4200, Image: "d"},
	}
	out := r.Rank(item, products, "male", "Casual")
	for _, p := range out {
		low := strings.ToLower(p.Title)
		if strings.Contains(low, "women") || strings.Contains(low, "corset") {
			t.Fatalf("unexpected product: %s", p.Title)
		}
	}
	if len(out) < 1 {
		t.Fatal("expected at least one men's tank")
	}
}

func TestRejectTankRequiresSleeveless(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Tank Top", Color: "White", Pattern: "Solid", Fit: "Slim", Material: "Cotton"}
	products := []models.Product{
		{Title: "Men White Slim Fit Ribbed Tank Top", Brand: "Westside", Website: "westside", Price: 499, Image: "a"},
		{Title: "Men White Slim Fit Cotton T-Shirt", Brand: "Westside", Website: "westside", Price: 499, Image: "b"},
		{Title: "Men White Oversized Graphic Tee", Brand: "Bluorng", Website: "bluorng", Price: 3700, Image: "c"},
		{Title: "Brown Long Sleeve Crew Neck", Brand: "Bluorng", Website: "bluorng", Price: 3000, Image: "d"},
	}
	out := r.Rank(item, products, "male", "Casual")
	if len(out) != 1 || !strings.Contains(strings.ToLower(out[0].Title), "tank") {
		t.Fatalf("expected only tank top, got %#v", out)
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
	out := r.Rank(item, products, "male", "Casual")
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

func TestRejectTieDyeForNecktie(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Tie", Color: "Black", Pattern: "Solid", Fit: "Regular", Material: "Silk"}
	products := []models.Product{
		{Title: "Alvaro Castagnino Men Black Necktie", Brand: "Alvaro", Website: "myntra", Price: 617, Image: "a"},
		{Title: "Amplify Men Black Tie and Dye T-Shirt", Brand: "Amplify", Website: "bearhouse", Price: 899, Image: "b"},
		{Title: "Bombay Paisley Black Dress", Brand: "Bombay Paisley", Website: "westside", Price: 1999, Image: "c"},
		{Title: "Powerlook Black Tie Dye Knit Polo", Brand: "Powerlook", Website: "powerlook", Price: 1299, Image: "d"},
	}
	out := r.Rank(item, products, "male", "Formal")
	if len(out) != 1 {
		t.Fatalf("expected 1 necktie, got %d %#v", len(out), out)
	}
	if !strings.Contains(strings.ToLower(out[0].Title), "necktie") && !strings.Contains(strings.ToLower(out[0].Title), "tie") {
		t.Fatalf("unexpected: %s", out[0].Title)
	}
	if strings.Contains(strings.ToLower(out[0].Title), "dye") {
		t.Fatalf("got tie-dye: %s", out[0].Title)
	}
}

func TestRejectTieAndDyeVariant(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Tie", Color: "Black", Pattern: "Solid", Material: "Silk"}
	products := []models.Product{
		{Title: "Men Black Formal Necktie", Brand: "Van Heusen", Website: "myntra", Price: 799, Image: "a"},
		{Title: "Amplify Men Black Tie and Dye T-Shirt", Brand: "Amplify", Website: "bearhouse", Price: 899, Image: "b"},
	}
	out := r.Rank(item, products, "male", "Formal")
	if len(out) != 1 || !strings.Contains(strings.ToLower(out[0].Title), "necktie") {
		t.Fatalf("expected necktie only, got %#v", out)
	}
}

func TestRejectEmbroideredFormalShirt(t *testing.T) {
	r := NewScoreRanker()
	item := models.ClothingItem{Category: "Shirt", Color: "White", Pattern: "Solid", Fit: "Regular", Material: "Cotton"}
	products := []models.Product{
		{Title: "Men White Cotton Formal Dress Shirt", Brand: "Roadster", Website: "myntra", Price: 574, Image: "a"},
		{Title: "Powerlook White Slub Embroidered Shirt", Brand: "Powerlook", Website: "powerlook", Price: 1299, Image: "b"},
		{Title: "Gia White Embroidered Cotton Shirt", Brand: "Gia", Website: "westside", Price: 1499, Image: "c"},
	}
	out := r.Rank(item, products, "male", "Formal")
	if len(out) != 1 {
		t.Fatalf("expected 1 formal shirt, got %d %#v", len(out), out)
	}
}
