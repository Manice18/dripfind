package storage

import (
	"context"
	"fmt"
	"strings"
)

// BackendFromConfig builds local or S3 storage from the given settings.
func BackendFromConfig(ctx context.Context, backend string, local LocalConfig, s3cfg S3Config) (ImageStorage, error) {
	switch strings.ToLower(strings.TrimSpace(backend)) {
	case "", "local":
		return NewLocalStorage(local.Dir, local.PublicBase)
	case "s3":
		return NewS3Storage(ctx, s3cfg)
	default:
		return nil, fmt.Errorf("unknown STORAGE_BACKEND %q (use local or s3)", backend)
	}
}

type LocalConfig struct {
	Dir        string
	PublicBase string
}
