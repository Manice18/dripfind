package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	PublicBaseURL  string // e.g. http://localhost:9000/lookbook
	ForcePathStyle bool
	PresignTTL     time.Duration // used when PublicBaseURL is empty
}

type S3Storage struct {
	client     *minio.Client
	bucket     string
	publicBase string
	presignTTL time.Duration
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if err := requireNonEmpty("S3_BUCKET", cfg.Bucket); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("S3_ACCESS_KEY_ID", cfg.AccessKey); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("S3_SECRET_ACCESS_KEY", cfg.SecretKey); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("S3_ENDPOINT", cfg.Endpoint); err != nil {
		return nil, err
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.PresignTTL <= 0 {
		cfg.PresignTTL = time.Hour
	}

	endpoint, secure, err := parseEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
	}
	if cfg.ForcePathStyle {
		opts.BucketLookup = minio.BucketLookupPath
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	s := &S3Storage{
		client:     client,
		bucket:     cfg.Bucket,
		publicBase: strings.TrimRight(cfg.PublicBaseURL, "/"),
		presignTTL: cfg.PresignTTL,
	}

	if err := s.ensureBucket(ctx, cfg.Region); err != nil {
		return nil, err
	}
	if s.publicBase != "" {
		if err := s.ensurePublicRead(ctx); err != nil {
			return nil, fmt.Errorf("set bucket public-read policy: %w", err)
		}
	}
	return s, nil
}

func parseEndpoint(raw string) (host string, secure bool, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("empty S3_ENDPOINT")
	}
	if !strings.Contains(raw, "://") {
		return strings.TrimRight(raw, "/"), false, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false, fmt.Errorf("parse S3_ENDPOINT: %w", err)
	}
	if u.Host == "" {
		return "", false, fmt.Errorf("S3_ENDPOINT missing host")
	}
	return u.Host, u.Scheme == "https", nil
}

func (s *S3Storage) ensureBucket(ctx context.Context, region string) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("bucket exists %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: region})
	if err != nil {
		// Race: another process created it.
		exists, existsErr := s.client.BucketExists(ctx, s.bucket)
		if existsErr == nil && exists {
			return nil
		}
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *S3Storage) ensurePublicRead(ctx context.Context) error {
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, s.bucket)
	return s.client.SetBucketPolicy(ctx, s.bucket, policy)
}

func (s *S3Storage) Save(ctx context.Context, data []byte, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	key := newObjectKey(contentType)
	if err := s.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return "", err
	}
	return key, nil
}

// Put writes an object at an exact key (used by Save and image migration).
func (s *S3Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	key = normalizeKey(key)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	key = normalizeKey(key)
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	// Stat forces the error for missing keys (GetObject is lazy).
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}

func (s *S3Storage) PublicURL(key string) string {
	key = normalizeKey(key)
	if key == "" || key == "." {
		return ""
	}
	if s.publicBase != "" {
		return joinURL(s.publicBase, key)
	}
	u, err := s.client.PresignedGetObject(context.Background(), s.bucket, key, s.presignTTL, nil)
	if err != nil {
		return ""
	}
	return u.String()
}

var _ ImageStorage = (*S3Storage)(nil)
