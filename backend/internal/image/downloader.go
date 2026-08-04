package image

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/manice18/dripfind/backend/internal/safehttp"
)

type Downloader struct {
	client *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout:   45 * time.Second,
			Transport: safehttp.Transport(),
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

type DownloadResult struct {
	Data        []byte
	ContentType string
}

func (d *Downloader) Download(imageURL string) (*DownloadResult, error) {
	req, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Dripfind/1.0)")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 25<<20))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image body")
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || !isAllowedImageType(ct) {
		ct = http.DetectContentType(data)
	}
	if !isAllowedImageType(ct) {
		return nil, fmt.Errorf("url did not return an allowed image type (content-type=%s)", ct)
	}

	return &DownloadResult{Data: data, ContentType: ct}, nil
}

func isAllowedImageType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}
