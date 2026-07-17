package search

import (
	"strings"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

// BuildQuery turns clothing attributes into a deterministic shopping query.
// The AI must not generate search strings — this function owns that.
func BuildQuery(item models.ClothingItem, gender, occasion string) string {
	parts := make([]string, 0, 8)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		lower := strings.ToLower(s)
		if lower == "other" || lower == "unknown" || lower == "solid" {
			return
		}
		for _, p := range parts {
			if strings.EqualFold(p, s) {
				return
			}
		}
		parts = append(parts, s)
	}

	add(primaryColor(item.Color))
	// Skip pattern noise for ties/suits (solid is default).
	cat := strings.ToLower(strings.TrimSpace(item.Category))
	if cat != "tie" && cat != "suit" && cat != "blazer" {
		add(item.Pattern)
	}
	add(fitSearchTerm(item))
	add(materialSearchTerm(item))
	add(categorySearchTerm(item, occasion))

	switch strings.ToLower(strings.TrimSpace(gender)) {
	case "male", "men", "man":
		parts = append([]string{"Men"}, parts...)
	case "female", "women", "woman":
		parts = append([]string{"Women"}, parts...)
	}

	if len(parts) == 0 {
		return "clothing"
	}
	return strings.Join(parts, " ")
}

func BuildQueryFromVision(item models.VisionClothingItem, gender, occasion string) string {
	return BuildQuery(models.ClothingItem{
		Category: item.Category,
		Color:    item.Color,
		Material: item.Material,
		Fit:      item.Fit,
		Pattern:  item.Pattern,
	}, gender, occasion)
}

func primaryColor(color string) string {
	color = strings.TrimSpace(color)
	lower := strings.ToLower(color)
	for _, sep := range []string{" and ", "/", ",", "&", "+"} {
		if i := strings.Index(lower, sep); i > 0 {
			return strings.TrimSpace(color[:i])
		}
	}
	return color
}

func fitSearchTerm(item models.ClothingItem) string {
	fit := strings.ToLower(strings.TrimSpace(item.Fit))
	cat := strings.ToLower(strings.TrimSpace(item.Category))
	if fit == "" {
		return ""
	}
	if cat == "tie" || cat == "suit" || cat == "blazer" {
		return ""
	}
	if (cat == "trousers" || cat == "jeans") && (fit == "relaxed" || fit == "wide" || fit == "oversized") {
		return "Baggy"
	}
	if fit == "oversized" {
		return "Oversized"
	}
	return strings.TrimSpace(item.Fit)
}

func materialSearchTerm(item models.ClothingItem) string {
	mat := strings.ToLower(strings.TrimSpace(item.Material))
	cat := strings.ToLower(strings.TrimSpace(item.Category))
	if strings.Contains(mat, "denim") || cat == "jeans" {
		return "Denim"
	}
	if cat == "tie" && (mat == "" || mat == "other") {
		return "Silk"
	}
	if mat == "" || mat == "other" || mat == "unknown" {
		return ""
	}
	return strings.TrimSpace(item.Material)
}

func categorySearchTerm(item models.ClothingItem, occasion string) string {
	cat := strings.ToLower(strings.TrimSpace(item.Category))
	mat := strings.ToLower(strings.TrimSpace(item.Material))
	fit := strings.ToLower(strings.TrimSpace(item.Fit))
	occ := strings.ToLower(strings.TrimSpace(occasion))
	formal := occ == "formal" || occ == "business"

	switch cat {
	case "trousers":
		if strings.Contains(mat, "denim") {
			return "Jeans"
		}
		if formal {
			return "Formal Trousers"
		}
		if fit == "relaxed" || fit == "wide" || fit == "oversized" {
			return "Trousers"
		}
		return "Casual Trousers"
	case "jeans":
		return "Jeans"
	case "shirt":
		if formal {
			return "Formal Dress Shirt"
		}
		return "Casual Shirt"
	case "t-shirt", "tshirt":
		return "Plain T-Shirt"
	case "tank top", "tanktop", "vest":
		if strings.Contains(mat, "rib") {
			return "Ribbed Tank Top"
		}
		return "Tank Top Sleeveless"
	case "tie":
		return "Formal Necktie"
	case "suit":
		return "Formal Suit"
	case "blazer":
		if formal {
			return "Formal Blazer"
		}
		return "Blazer"
	case "watch":
		return "Watch"
	case "bag":
		return "Backpack"
	case "jewelry":
		return "Chain Necklace"
	case "sunglasses":
		return "Sunglasses"
	case "belt":
		if formal {
			return "Formal Leather Belt"
		}
		return "Belt"
	default:
		return strings.TrimSpace(item.Category)
	}
}

func EnsureProductDefaults(p *models.Product) {
	if p.Currency == "" {
		p.Currency = "INR"
	}
	if p.Brand == "" {
		p.Brand = "Unknown"
	}
	if p.Title == "" {
		p.Title = p.Website + " item"
	}
}
