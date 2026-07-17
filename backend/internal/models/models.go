package models

import (
	"time"

	"github.com/google/uuid"
)

type OutfitStatus string

const (
	StatusPending    OutfitStatus = "pending"
	StatusProcessing OutfitStatus = "processing"
	StatusCompleted  OutfitStatus = "completed"
	StatusFailed     OutfitStatus = "failed"
)

type ClothingItem struct {
	ID          uuid.UUID `json:"id"`
	OutfitID    uuid.UUID `json:"outfit_id,omitempty"`
	Category    string    `json:"category"`
	Color       string    `json:"color"`
	Material    string    `json:"material"`
	Fit         string    `json:"fit"`
	Pattern     string    `json:"pattern"`
	Confidence  float64   `json:"confidence"`
	SearchQuery string    `json:"search_query,omitempty"`
	Products    []Product `json:"products,omitempty"`
}

type Product struct {
	ID         uuid.UUID `json:"id"`
	ItemID     uuid.UUID `json:"item_id,omitempty"`
	Title      string    `json:"title"`
	Brand      string    `json:"brand"`
	Price      float64   `json:"price"`
	Currency   string    `json:"currency"`
	Website    string    `json:"website"`
	URL        string    `json:"url"`
	Image      string    `json:"image"`
	MatchScore float64   `json:"match_score"`
}

type Outfit struct {
	ID           uuid.UUID      `json:"id"`
	ImagePath    string         `json:"image_path"`
	ImageURL     string         `json:"image_url,omitempty"`
	SourceURL    string         `json:"source_url"`
	Style        string         `json:"style"`
	Gender       string         `json:"gender"`
	Season       string         `json:"season"`
	Occasion     string         `json:"occasion"`
	Status       OutfitStatus   `json:"status"`
	ErrorMessage string         `json:"error_message,omitempty"`
	Items        []ClothingItem `json:"items,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type HistoryEntry struct {
	ID        uuid.UUID `json:"id"`
	OutfitID  uuid.UUID `json:"outfit_id"`
	SourceURL string    `json:"source_url"`
	Style     string    `json:"style,omitempty"`
	Gender    string    `json:"gender,omitempty"`
	Status    string    `json:"status,omitempty"`
	ImageURL  string    `json:"image_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// VisionOutfit is the structured response from the vision analyzer.
type VisionOutfit struct {
	Gender   string             `json:"gender"`
	Style    string             `json:"style"`
	Season   string             `json:"season"`
	Occasion string             `json:"occasion"`
	Items    []VisionClothingItem `json:"items"`
}

type VisionClothingItem struct {
	Category   string  `json:"category"`
	Color      string  `json:"color"`
	Material   string  `json:"material"`
	Fit        string  `json:"fit"`
	Pattern    string  `json:"pattern"`
	Confidence float64 `json:"confidence"`
}
