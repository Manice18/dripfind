package ranking

import (
	"math"
	"sort"
	"strings"

	"github.com/manice18/outfit_finder/backend/internal/models"
)

type Ranker interface {
	Rank(item models.ClothingItem, products []models.Product, gender, occasion string) []models.Product
}

type ScoreRanker struct {
	MaxPrice      float64
	MinKeepScore  float64
	MaxPerWebsite int
	Limit         int
}

func NewScoreRanker() *ScoreRanker {
	return &ScoreRanker{
		MaxPrice:      15000,
		MinKeepScore:  55,
		MaxPerWebsite: 2,
		Limit:         12,
	}
}

func (r *ScoreRanker) Rank(item models.ClothingItem, products []models.Product, gender, occasion string) []models.Product {
	if len(products) == 0 {
		return products
	}

	primary := make([]models.Product, 0, len(products))
	backup := make([]models.Product, 0, len(products))
	seen := map[string]struct{}{}

	for _, p := range products {
		key := dedupeKey(p)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		if hardReject(item, p, gender, occasion) {
			continue
		}

		p.MatchScore = r.score(item, p, gender)
		if p.MatchScore >= r.MinKeepScore {
			primary = append(primary, p)
		} else if p.MatchScore >= r.MinKeepScore-10 {
			backup = append(backup, p)
		}
	}

	sort.SliceStable(primary, func(i, j int) bool {
		if primary[i].MatchScore == primary[j].MatchScore {
			return primary[i].Price < primary[j].Price
		}
		return primary[i].MatchScore > primary[j].MatchScore
	})
	sort.SliceStable(backup, func(i, j int) bool {
		return backup[i].MatchScore > backup[j].MatchScore
	})

	merged := primary
	// If filters left the grid too empty, fill with next-best safe matches.
	if len(merged) < 6 {
		merged = append(merged, backup...)
	}

	return diversify(merged, r.MaxPerWebsite, r.Limit)
}

// diversify keeps score order but caps how many results come from one site,
// so one retailer can't fill the whole grid.
func diversify(products []models.Product, maxPerWebsite, limit int) []models.Product {
	if maxPerWebsite <= 0 {
		maxPerWebsite = 2
	}
	if limit <= 0 {
		limit = 10
	}

	out := make([]models.Product, 0, limit)
	counts := map[string]int{}

	for _, p := range products {
		if len(out) >= limit {
			break
		}
		site := strings.ToLower(p.Website)
		if counts[site] >= maxPerWebsite {
			continue
		}
		counts[site]++
		out = append(out, p)
	}

	// Second pass: if still sparse, only add from sites not yet shown.
	if len(out) < 6 {
		for _, p := range products {
			if len(out) >= limit {
				break
			}
			site := strings.ToLower(p.Website)
			if counts[site] > 0 {
				continue
			}
			if alreadyIn(out, p) {
				continue
			}
			counts[site]++
			out = append(out, p)
		}
	}

	return out
}

func alreadyIn(out []models.Product, p models.Product) bool {
	for _, o := range out {
		if o.Image == p.Image && o.Title == p.Title {
			return true
		}
	}
	return false
}

func hardReject(item models.ClothingItem, p models.Product, gender, occasion string) bool {
	title := strings.ToLower(p.Title + " " + p.Brand + " " + p.Website)

	if isKidsOrBaby(title) {
		return true
	}
	if isCategoryMismatch(title, item.Category) {
		return true
	}
	if genderMismatch(title, gender) {
		return true
	}

	// Striped/checked/etc must appear in the title when vision says so.
	pattern := strings.ToLower(strings.TrimSpace(item.Pattern))
	cat := strings.ToLower(strings.TrimSpace(item.Category))
	if cat != "tie" && pattern != "" && pattern != "solid" && pattern != "other" {
		if !patternSynonym(title, item.Pattern) && !containsAny(title, item.Pattern) {
			return true
		}
	}

	// Denim bottoms must be denim/jeans — not chino/formal.
	if wantsDenim(item) {
		if !isDenimTitle(title) {
			return true
		}
		if isFormalBottom(title) {
			return true
		}
	}

	// Reject clear color conflicts, but allow titles that omit the color word.
	if item.Color != "" && hasConflictingColor(title, item.Color) && !colorMatch(title, item.Color) {
		return true
	}
	if hasLoudConflictColor(title, item.Color) && !colorMatch(title, item.Color) {
		return true
	}

	// Solid looks should not return graphic/print-heavy products (except ties with subtle weave).
	if strings.EqualFold(item.Pattern, "Solid") && cat != "tie" && isGraphicTitle(title) {
		return true
	}

	if isFormalOccasion(occasion) && rejectsFormalLook(title, cat) {
		return true
	}

	fit := strings.ToLower(item.Fit)
	if (cat == "jeans" || cat == "trousers") && (fit == "relaxed" || fit == "oversized" || fit == "wide") {
		if containsAny(title, "skinny") || containsAny(title, "super slim") {
			return true
		}
	}

	return false
}

func isFormalOccasion(occasion string) bool {
	o := strings.ToLower(strings.TrimSpace(occasion))
	return o == "formal" || o == "business"
}

func rejectsFormalLook(title, category string) bool {
	switch strings.ToLower(category) {
	case "shirt":
		return containsAny(title, "embroider") || containsAny(title, "casual") ||
			containsAny(title, "oversized") || containsAny(title, "graphic") ||
			containsAny(title, "tie dye") || containsAny(title, "tie-dye")
	case "suit", "blazer":
		return containsAny(title, "tote") || containsAny(title, "bag") ||
			containsAny(title, "backpack") || containsAny(title, "duffle")
	default:
		return false
	}
}

func isKidsOrBaby(title string) bool {
	markers := []string{
		"baby", "infant", "toddler", "newborn", "kids", "kid's", "kids'",
		"boys ", "boy's", "girls ", "girl's", "junior", "children", "child ",
		"hop baby", "0-3", "3-6", "6-9", "9-12 months", "years kids",
	}
	for _, m := range markers {
		if strings.Contains(title, m) {
			return true
		}
	}
	return false
}

func hasLoudConflictColor(title, want string) bool {
	wantParts := splitColors(want)
	wantSet := map[string]struct{}{}
	for _, w := range wantParts {
		wantSet[strings.ToLower(w)] = struct{}{}
		for _, s := range colorSynonyms(w) {
			wantSet[s] = struct{}{}
		}
	}
	loud := []string{"pink", "red", "green", "yellow", "purple", "orange", "maroon", "burgundy", "dusty pink"}
	for _, c := range loud {
		if _, ok := wantSet[c]; ok {
			continue
		}
		if strings.Contains(title, c) {
			return true
		}
	}
	return false
}

func (r *ScoreRanker) score(item models.ClothingItem, p models.Product, gender string) float64 {
	title := strings.ToLower(p.Title + " " + p.Brand)
	var points float64
	var weight float64

	add := func(w float64, hit bool) {
		weight += w
		if hit {
			points += w
		}
	}

	colorHit := colorMatch(title, item.Color)
	catHit := categoryMatch(title, item.Category)
	matHit := materialMatch(title, item.Material)
	fitHit := fitMatch(title, item.Fit)
	patHit := true
	if !strings.EqualFold(item.Pattern, "Solid") && item.Pattern != "" {
		patHit = patternSynonym(title, item.Pattern) || containsAny(title, item.Pattern)
	}

	add(0.28, colorHit)
	add(0.22, catHit)
	add(0.16, matHit)
	add(0.12, fitHit)
	add(0.14, patHit)
	add(0.04, p.Brand != "" && !strings.EqualFold(p.Brand, "Unknown"))
	add(0.04, genderAligned(title, gender))

	if item.Color != "" && !colorHit && hasConflictingColor(title, item.Color) {
		points -= 0.35
	}
	if !catHit {
		points -= 0.18
	}
	if wantsDenim(item) && isDenimTitle(title) {
		points += 0.08
		weight += 0.08
	}

	priceScore := 0.0
	if p.Price > 0 && r.MaxPrice > 0 {
		ratio := p.Price / r.MaxPrice
		if ratio > 1 {
			ratio = 1
		}
		priceScore = 1 - math.Abs(ratio-0.2)/0.8
		if priceScore < 0 {
			priceScore = 0
		}
	}
	weight += 0.04
	points += 0.04 * priceScore

	if weight <= 0 {
		return 30
	}
	pct := (points / weight) * 100
	if pct < 0 {
		pct = 0
	}
	if pct > 98 {
		pct = 98
	}
	return math.Round(pct*10) / 10
}

func wantsDenim(item models.ClothingItem) bool {
	mat := strings.ToLower(item.Material)
	cat := strings.ToLower(item.Category)
	return strings.Contains(mat, "denim") || cat == "jeans"
}

func isDenimTitle(title string) bool {
	return containsAny(title, "denim") || containsAny(title, "jeans") || containsAny(title, "jean")
}

func isFormalBottom(title string) bool {
	return containsAny(title, "formal") || containsAny(title, "pleated") || containsAny(title, "dress pant") ||
		containsAny(title, "office") || containsAny(title, "chino")
}

func genderMismatch(title, gender string) bool {
	g := strings.ToLower(strings.TrimSpace(gender))
	switch g {
	case "male", "men", "man":
		if containsAny(title, "women") || containsAny(title, "woman") || containsAny(title, "ladies") ||
			containsAny(title, "girls") || strings.Contains(title, "womens") || strings.Contains(title, "womenswear") {
			return true
		}
		// Female-coded garment types / brands that rarely appear on men's listings.
		femaleCoded := []string{
			"corset", "bralette", "crop top", "cropped top", "blouse", "skort",
			"bodycon", "camisole", "cami ", "lingerie", "shapewear", "saree",
			"lehenga", "kurti", "anarkali", "peplum",
			" lov ", "lov ", " gia ", "gia ", " nuon ", "nuon ",
			"bombay paisley", "a-line dress", "plus size",
		}
		for _, w := range femaleCoded {
			if strings.Contains(title, w) {
				return true
			}
		}
		return false
	case "female", "women", "woman":
		return containsAny(title, "men's") || containsAny(title, "mens ") || strings.HasPrefix(title, "men ") ||
			(containsAny(title, "male") && !containsAny(title, "female"))
	default:
		return false
	}
}

func hasMensSignal(title string) bool {
	return strings.Contains(title, "men's") || strings.Contains(title, "mens ") ||
		strings.Contains(title, "menswear") || strings.Contains(title, " for men") ||
		strings.Contains(title, "- men") || strings.Contains(title, " men ") ||
		strings.HasPrefix(title, "men ") || strings.Contains(title, "male ") ||
		strings.Contains(title, " men-") || strings.HasSuffix(title, " men")
}

func genderAligned(title, gender string) bool {
	g := strings.ToLower(strings.TrimSpace(gender))
	switch g {
	case "male", "men", "man":
		return containsAny(title, "men") || containsAny(title, "male") || containsAny(title, "mens")
	case "female", "women", "woman":
		return containsAny(title, "women") || containsAny(title, "woman") || containsAny(title, "ladies")
	default:
		return true
	}
}

func colorMatch(title, color string) bool {
	color = strings.ToLower(strings.TrimSpace(color))
	if color == "" {
		return false
	}
	// Support compound colors like "White and gray"
	for _, part := range splitColors(color) {
		if containsAny(title, part) {
			return true
		}
		for _, syn := range colorSynonyms(part) {
			if strings.Contains(title, syn) {
				return true
			}
		}
	}
	return false
}

func splitColors(color string) []string {
	color = strings.ToLower(color)
	for _, sep := range []string{" and ", "/", ",", "&", "+"} {
		if strings.Contains(color, sep) {
			parts := strings.Split(color, sep)
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return []string{color}
}

func colorSynonyms(color string) []string {
	switch strings.ToLower(color) {
	case "beige", "tan", "khaki", "sand":
		return []string{"beige", "tan", "khaki", "cream", "sand", "stone"}
	case "cream", "off-white", "off white", "ivory", "white":
		return []string{"cream", "ivory", "off white", "off-white", "white"}
	case "black":
		return []string{"black", "noir", "charcoal"}
	case "navy", "blue":
		return []string{"navy", "blue"}
	case "brown":
		return []string{"brown", "cognac", "chocolate"}
	case "grey", "gray":
		return []string{"grey", "gray", "charcoal"}
	default:
		return nil
	}
}

func hasConflictingColor(title, want string) bool {
	wantParts := splitColors(want)
	wantSet := map[string]struct{}{}
	for _, w := range wantParts {
		wantSet[w] = struct{}{}
		for _, s := range colorSynonyms(w) {
			wantSet[s] = struct{}{}
		}
	}

	conflicts := []string{"red", "blue", "green", "yellow", "pink", "purple", "orange", "maroon", "burgundy", "dusty pink"}
	for _, c := range conflicts {
		if _, ok := wantSet[c]; ok {
			continue
		}
		if strings.Contains(title, c) {
			return true
		}
	}
	return false
}

func materialMatch(title, material string) bool {
	material = strings.ToLower(strings.TrimSpace(material))
	if material == "" {
		return false
	}
	if containsAny(title, material) {
		return true
	}
	if strings.Contains(material, "denim") {
		return isDenimTitle(title)
	}
	if strings.Contains(material, "linen") {
		return containsAny(title, "linen")
	}
	if strings.Contains(material, "cotton") {
		return containsAny(title, "cotton")
	}
	return false
}

func fitMatch(title, fit string) bool {
	fit = strings.ToLower(strings.TrimSpace(fit))
	if fit == "" {
		return false
	}
	if containsAny(title, fit) {
		return true
	}
	switch fit {
	case "relaxed", "oversized", "wide":
		return containsAny(title, "baggy") || containsAny(title, "loose") || containsAny(title, "wide") ||
			containsAny(title, "oversized") || containsAny(title, "relaxed") || containsAny(title, "boxy")
	case "slim":
		return containsAny(title, "slim") || containsAny(title, "skinny")
	default:
		return false
	}
}

func categoryMatch(title, category string) bool {
	category = strings.ToLower(strings.TrimSpace(category))
	if isCategoryMismatch(title, category) {
		return false
	}
	if containsAny(title, category) {
		return true
	}
	switch category {
	case "trousers":
		return containsAny(title, "pants") || containsAny(title, "trousers") || containsAny(title, "jeans")
	case "jeans":
		return isDenimTitle(title)
	case "shirt":
		return containsAny(title, "shirt") && !isTShirtTitle(title) && !containsAny(title, "polo") && !isTankTitle(title)
	case "t-shirt", "tshirt":
		return isTShirtTitle(title) || (containsAny(title, "tee") && !isTankTitle(title))
	case "tank top", "tanktop", "vest":
		return isTankTitle(title)
	case "tie":
		return isNecktieTitle(title)
	case "suit":
		return containsAny(title, "suit") && !containsAny(title, "swimsuit") && !containsAny(title, "tracksuit")
	case "blazer":
		return containsAny(title, "blazer") || containsAny(title, "suit jacket")
	case "bag":
		return containsAny(title, "backpack") || containsAny(title, "bag")
	case "watch":
		return containsAny(title, "watch")
	case "jewelry":
		return containsAny(title, "necklace") || containsAny(title, "chain") || containsAny(title, "jewellery") || containsAny(title, "jewelry")
	case "sunglasses":
		return containsAny(title, "sunglass") || containsAny(title, "eyewear")
	case "belt":
		return containsAny(title, "belt")
	default:
		return false
	}
}

func isCategoryMismatch(title, category string) bool {
	category = strings.ToLower(strings.TrimSpace(category))
	switch category {
	case "shirt":
		if isTShirtTitle(title) || containsAny(title, "polo") || isTankTitle(title) {
			return true
		}
		// "dress shirt" is valid; bare "dress" (womenswear) is not.
		if containsAny(title, "dress") && !containsAny(title, "dress shirt") {
			return true
		}
		return containsAny(title, "kurta") || containsAny(title, "salwar") || containsAny(title, "sherwani") ||
			containsAny(title, "pyjama") || containsAny(title, "pajama") ||
			(containsAny(title, "set") && (containsAny(title, "pyjama") || containsAny(title, "pajama")))
	case "t-shirt", "tshirt":
		if isTankTitle(title) {
			return true
		}
		if containsAny(title, "shirt") && !isTShirtTitle(title) && !strings.Contains(title, "tshirt") {
			if !strings.Contains(title, "t-shirt") && !strings.Contains(title, "tee") {
				return containsAny(title, "casual shirt") || containsAny(title, "linen shirt") || containsAny(title, "oxford")
			}
		}
		return false
	case "tank top", "tanktop", "vest":
		if isTShirtTitle(title) && !strings.Contains(title, "sleeveless") {
			return true
		}
		if containsAny(title, "round neck") || containsAny(title, "polo") || containsAny(title, "henley") {
			return true
		}
		if containsAny(title, "sleeve") && !containsAny(title, "sleeveless") {
			return true
		}
		if !isTankTitle(title) {
			return true
		}
		return false
	case "tie":
		// "Tie" must mean necktie — never tie-dye / tie-up apparel.
		if isTieDyeOrTieUp(title) {
			return true
		}
		if isTShirtTitle(title) || containsAny(title, "dress") || containsAny(title, "polo") ||
			containsAny(title, "hoodie") || containsAny(title, "jeans") ||
			(containsAny(title, "shirt") && !isNecktieTitle(title)) {
			return true
		}
		return !isNecktieTitle(title)
	case "suit":
		if containsAny(title, "bag") || containsAny(title, "tote") || containsAny(title, "backpack") ||
			containsAny(title, "tracksuit") || containsAny(title, "swimsuit") || isTShirtTitle(title) {
			return true
		}
		if !containsAny(title, "suit") && !containsAny(title, "blazer") {
			return true
		}
		return false
	case "blazer":
		return !containsAny(title, "blazer") && !containsAny(title, "suit jacket")
	case "trousers", "jeans":
		return containsAny(title, "skirt") || containsAny(title, "shorts") || containsAny(title, "legging") ||
			containsAny(title, "track pant")
	case "watch":
		return containsAny(title, "shirt") || containsAny(title, "shoe") || containsAny(title, "bag")
	default:
		return false
	}
}

func isTieDyeOrTieUp(title string) bool {
	return strings.Contains(title, "tie-dye") || strings.Contains(title, "tie dye") ||
		strings.Contains(title, "tiedye") || strings.Contains(title, "tie and dye") ||
		strings.Contains(title, "tie-up") || strings.Contains(title, "tie up")
}

func isNecktieTitle(title string) bool {
	if isTieDyeOrTieUp(title) {
		return false
	}
	if containsAny(title, "necktie") || containsAny(title, "neck tie") || containsAny(title, "cravat") {
		return true
	}
	looksLikeTie := strings.Contains(title, " tie") || strings.HasSuffix(title, "tie") || strings.Contains(title, "tie ")
	if !looksLikeTie {
		return false
	}
	return !isTShirtTitle(title) && !containsAny(title, "shirt") && !containsAny(title, "dress") && !containsAny(title, "polo")
}

func isTankTitle(title string) bool {
	// Prefer explicit tank/sleeveless. Bare "vest" is OK for Indian men's undershirts.
	if containsAny(title, "tank") || containsAny(title, "sleeveless") ||
		containsAny(title, "muscle tee") || containsAny(title, "ribbed tank") {
		return true
	}
	if containsAny(title, "vest") && !isTShirtTitle(title) {
		return true
	}
	return false
}

func isGraphicTitle(title string) bool {
	return containsAny(title, "graphic") || containsAny(title, "print") || containsAny(title, "logo") ||
		containsAny(title, "floral") || containsAny(title, "typography") || containsAny(title, "printed")
}

func isTShirtTitle(title string) bool {
	compact := strings.ReplaceAll(title, "-", "")
	compact = strings.ReplaceAll(compact, " ", "")
	if strings.Contains(compact, "tshirt") {
		return true
	}
	return strings.Contains(title, "t-shirt") || strings.Contains(title, "tee ") ||
		strings.HasSuffix(title, " tee") || strings.Contains(title, " tee")
}

func patternSynonym(title, pattern string) bool {
	switch strings.ToLower(pattern) {
	case "striped":
		return strings.Contains(title, "stripe") || strings.Contains(title, "pinstripe") || strings.Contains(title, "striped")
	case "checked":
		return strings.Contains(title, "check") || strings.Contains(title, "plaid") || strings.Contains(title, "gingham")
	default:
		return false
	}
}

func containsAny(haystack, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	if needle == "" || needle == "other" {
		return false
	}
	parts := strings.Fields(needle)
	for _, p := range parts {
		if len(p) < 2 {
			continue
		}
		if strings.Contains(haystack, p) {
			return true
		}
	}
	return strings.Contains(haystack, needle)
}

func dedupeKey(p models.Product) string {
	title := strings.ToLower(strings.Join(strings.Fields(p.Title), " "))
	if p.Image != "" {
		return p.Website + "|" + p.Image
	}
	return p.Website + "|" + title
}
