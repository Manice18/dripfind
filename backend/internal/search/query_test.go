package search

import (
	"testing"

	"github.com/manice18/dripfind/backend/internal/models"
)

func TestBuildQueryDeterministic(t *testing.T) {
	item := models.ClothingItem{
		Category: "Shirt",
		Color:    "White",
		Material: "Linen",
		Fit:      "Oversized",
		Pattern:  "Solid",
	}
	got := BuildQuery(item, "male", "Casual")
	want := "Men White Oversized Linen Casual Shirt"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildQueryFormalShirt(t *testing.T) {
	item := models.ClothingItem{
		Category: "Shirt",
		Color:    "White",
		Material: "Cotton",
		Fit:      "Regular",
		Pattern:  "Solid",
	}
	got := BuildQuery(item, "male", "Formal")
	want := "Men White Regular Cotton Formal Dress Shirt"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildQueryNecktie(t *testing.T) {
	item := models.ClothingItem{
		Category: "Tie",
		Color:    "Black",
		Material: "Silk",
		Fit:      "Regular",
		Pattern:  "Solid",
	}
	got := BuildQuery(item, "male", "Formal")
	want := "Men Black Silk Formal Necktie"
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
	got := BuildQuery(item, "male", "Casual")
	want := "Men Black Baggy Denim Jeans"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
