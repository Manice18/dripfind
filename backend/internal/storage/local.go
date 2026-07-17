package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type ImageStorage interface {
	Save(data []byte, contentType string) (string, error)
	AbsPath(rel string) string
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	originals := filepath.Join(baseDir, "originals")
	processed := filepath.Join(baseDir, "processed")
	for _, d := range []string{originals, processed} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

func (s *LocalStorage) Save(data []byte, contentType string) (string, error) {
	ext := extFromContentType(contentType)
	name := uuid.New().String() + ext
	rel := filepath.Join("originals", name)
	abs := filepath.Join(s.baseDir, rel)
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", fmt.Errorf("write image: %w", err)
	}
	return rel, nil
}

func (s *LocalStorage) AbsPath(rel string) string {
	return filepath.Join(s.baseDir, rel)
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
