package jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	KindURL    = "url"
	KindUpload = "upload"

	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusDead    = "dead"
)

type Job struct {
	ID          uuid.UUID
	OutfitID    uuid.UUID
	Kind        string
	Payload     json.RawMessage
	Status      string
	Attempts    int
	MaxAttempts int
	RunAt       time.Time
	LockedAt    *time.Time
	LastError   *string
	CreatedAt   time.Time
}

type URLPayload struct {
	URL string `json:"url"`
}

type UploadPayload struct {
	ImageKey    string `json:"image_key"`
	ContentType string `json:"content_type"`
	SourceLabel string `json:"source_label"`
}

func EncodePayload(v any) (json.RawMessage, error) {
	return json.Marshal(v)
}

func DecodeURLPayload(raw json.RawMessage) (URLPayload, error) {
	var p URLPayload
	err := json.Unmarshal(raw, &p)
	return p, err
}

func DecodeUploadPayload(raw json.RawMessage) (UploadPayload, error) {
	var p UploadPayload
	err := json.Unmarshal(raw, &p)
	return p, err
}

// Backoff returns how long to wait before the next attempt (1-based attempt number).
func Backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	// 2^attempt seconds: 2s, 4s, 8s, ...
	sec := 1 << attempt
	if sec > 300 {
		sec = 300
	}
	return time.Duration(sec) * time.Second
}
