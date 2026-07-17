package image

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Downloader struct {
	client *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 45 * time.Second,
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OutfitFinder/1.0)")
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
	if ct == "" {
		ct = http.DetectContentType(data)
	}
	if !strings.HasPrefix(ct, "image/") && !looksLikeImage(data) {
		return nil, fmt.Errorf("url did not return an image (content-type=%s)", ct)
	}
	if !strings.HasPrefix(ct, "image/") {
		ct = http.DetectContentType(data)
	}

	return &DownloadResult{Data: data, ContentType: ct}, nil
}

func looksLikeImage(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8:
		return true
	case data[0] == 0x89 && data[1] == 0x50:
		return true
	case data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
		return true
	case data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F':
		return true
	default:
		return false
	}
}
