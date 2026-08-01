package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage writes to a local directory and serves via /images/* on the API.
type LocalStorage struct {
	baseDir    string
	publicBase string // e.g. "/images"
}

func NewLocalStorage(baseDir, publicBase string) (*LocalStorage, error) {
	if publicBase == "" {
		publicBase = "/images"
	}
	originals := filepath.Join(baseDir, "originals")
	processed := filepath.Join(baseDir, "processed")
	for _, d := range []string{originals, processed} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	return &LocalStorage{baseDir: baseDir, publicBase: publicBase}, nil
}

func (s *LocalStorage) Save(_ context.Context, data []byte, contentType string) (string, error) {
	key := newObjectKey(contentType)
	abs := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", fmt.Errorf("write image: %w", err)
	}
	return key, nil
}

func (s *LocalStorage) Open(_ context.Context, key string) (io.ReadCloser, error) {
	key = normalizeKey(key)
	abs := filepath.Join(s.baseDir, filepath.FromSlash(key))
	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	return f, nil
}

func (s *LocalStorage) PublicURL(key string) string {
	if key == "" {
		return ""
	}
	return joinURL(s.publicBase, normalizeKey(key))
}

var _ ImageStorage = (*LocalStorage)(nil)
