package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/manice18/dripfind/backend/internal/models"
	"github.com/manice18/dripfind/backend/internal/prompt"
)

type Analyzer interface {
	Analyze(ctx context.Context, data []byte, contentType string) (*models.VisionOutfit, error)
}

type OpenAIAnalyzer struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIAnalyzer(apiKey, model string) *OpenAIAnalyzer {
	return &OpenAIAnalyzer{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

type DemoAnalyzer struct{}

func NewDemoAnalyzer() *DemoAnalyzer { return &DemoAnalyzer{} }

func (d *DemoAnalyzer) Analyze(_ context.Context, _ []byte, _ string) (*models.VisionOutfit, error) {
	return &models.VisionOutfit{
		Gender:   "male",
		Style:    "Old Money",
		Season:   "Summer",
		Occasion: "Casual",
		Items: []models.VisionClothingItem{
			{Category: "Shirt", Color: "White", Material: "Linen", Fit: "Oversized", Pattern: "Solid", Confidence: 0.96},
			{Category: "Trousers", Color: "Beige", Material: "Cotton", Fit: "Relaxed", Pattern: "Solid", Confidence: 0.94},
			{Category: "Shoes", Color: "Brown", Material: "Leather", Fit: "Regular", Pattern: "Solid", Confidence: 0.91},
			{Category: "Watch", Color: "Silver", Material: "Metal", Fit: "Regular", Pattern: "Solid", Confidence: 0.85},
		},
	}, nil
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (a *OpenAIAnalyzer) Analyze(ctx context.Context, data []byte, contentType string) (*models.VisionOutfit, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image")
	}
	if contentType == "" {
		contentType = "image/jpeg"
	}
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data))

	payload := chatRequest{
		Model: a.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: []contentPart{{Type: "text", Text: prompt.VisionSystem}},
			},
			{
				Role: "user",
				Content: []contentPart{
					{Type: "text", Text: prompt.VisionUser},
					{Type: "image_url", ImageURL: &imageURL{URL: dataURL}},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	content = stripCodeFence(content)

	var outfit models.VisionOutfit
	if err := json.Unmarshal([]byte(content), &outfit); err != nil {
		return nil, fmt.Errorf("parse vision json: %w\nraw: %s", err, truncate(content, 400))
	}
	if len(outfit.Items) == 0 {
		return nil, fmt.Errorf("vision returned no clothing items")
	}
	return &outfit, nil
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
