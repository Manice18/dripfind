package search

import (
	"testing"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

func TestBuildQueryDeterministic(t *testing.T) {
	item := models.ClothingItem{
		Category: "Shirt",
		Color:    "White",
		Material: "Linen",
		Fit:      "Oversized",
		Pattern:  "Solid",
	}
	got := BuildQuery(item, "male")
	want := "White Oversized Linen Casual Shirt Men"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildQueryStripedShirt(t *testing.T) {
	item := models.ClothingItem{
		Category: "Shirt",
		Color:    "White and gray",
		Material: "Cotton",
		Fit:      "Regular",
		Pattern:  "Striped",
	}
	got := BuildQuery(item, "male")
	want := "White Striped Regular Cotton Casual Shirt Men"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildQueryDenimTrousers(t *testing.T) {
	item := models.ClothingItem{
		Category: "Trousers",
		Color:    "Black",
		Material: "Denim",
		Fit:      "Relaxed",
		Pattern:  "Solid",
	}
	got := BuildQuery(item, "male")
	want := "Black Baggy Denim Jeans Men"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
