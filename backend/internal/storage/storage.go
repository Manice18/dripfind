package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/uuid"
)

// ImageStorage persists outfit images. Keys are opaque (e.g. "originals/<uuid>.jpg").
type ImageStorage interface {
	Save(ctx context.Context, data []byte, contentType string) (key string, err error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	PublicURL(key string) string
}

func extFromContentType(ct string) string {
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	switch ct {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

func ContentTypeFromKey(key string) string {
	switch strings.ToLower(path.Ext(key)) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

func newObjectKey(contentType string) string {
	return path.Join("originals", uuid.New().String()+extFromContentType(contentType))
}

func normalizeKey(key string) string {
	key = strings.TrimPrefix(key, "/")
	return path.Clean(key)
}

func joinURL(base, key string) string {
	base = strings.TrimRight(base, "/")
	key = strings.TrimLeft(key, "/")
	if base == "" {
		return "/" + key
	}
	return base + "/" + key
}

func requireNonEmpty(name, v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
