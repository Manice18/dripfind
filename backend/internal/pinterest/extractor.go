package pinterest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/manice18/dripfind/backend/internal/safehttp"
)

var (
	ogImageRe     = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	ogImageReAlt  = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["']`)
	pinimgRe      = regexp.MustCompile(`https://i\.pinimg\.com/[^"'\s]+`)
	directImageRe = regexp.MustCompile(`(?i)\.(jpe?g|png|webp|gif)(\?.*)?$`)
)

type Extractor struct {
	client *http.Client
}

func NewExtractor() *Extractor {
	return &Extractor{
		client: &http.Client{
			Timeout:   30 * time.Second,
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

func ValidateURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must be http or https")
	}
	if !IsPinterestHost(u.Host) {
		return fmt.Errorf("only Pinterest URLs are supported (pinterest.com or pin.it)")
	}
	return nil
}

func IsPinterestHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	switch host {
	case "pin.it", "www.pin.it", "i.pinimg.com",
		"pinterest.com", "www.pinterest.com":
		return true
	}
	// pinterest.co.uk, pinterest.ca, etc. — suffix match only.
	for _, root := range []string{"pinterest.com", "pinterest.co.uk", "pinterest.ca",
		"pinterest.de", "pinterest.fr", "pinterest.in", "pinterest.jp", "pinterest.au"} {
		if host == root || strings.HasSuffix(host, "."+root) {
			return true
		}
	}
	return false
}

func IsDirectPinImageURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	return host == "i.pinimg.com" && directImageRe.MatchString(u.Path)
}

// ExtractImageURL resolves a Pinterest pin URL into a downloadable image URL.
func (e *Extractor) ExtractImageURL(raw string) (string, error) {
	if err := ValidateURL(raw); err != nil {
		return "", err
	}
	raw = strings.TrimSpace(raw)

	if IsDirectPinImageURL(raw) {
		return preferLargePinImage(raw), nil
	}

	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch pin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("pinterest returned status %d", resp.StatusCode)
	}

	finalURL := resp.Request.URL.String()
	if IsDirectPinImageURL(finalURL) {
		return preferLargePinImage(finalURL), nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return "", err
	}
	html := string(body)

	if m := ogImageRe.FindStringSubmatch(html); len(m) > 1 {
		return preferLargePinImage(htmlUnescape(m[1])), nil
	}
	if m := ogImageReAlt.FindStringSubmatch(html); len(m) > 1 {
		return preferLargePinImage(htmlUnescape(m[1])), nil
	}
	if m := pinimgRe.FindString(html); m != "" {
		return preferLargePinImage(m), nil
	}

	return "", fmt.Errorf("could not extract image from this Pinterest URL")
}

func preferLargePinImage(imageURL string) string {
	imageURL = strings.Replace(imageURL, "/236x/", "/736x/", 1)
	imageURL = strings.Replace(imageURL, "/474x/", "/736x/", 1)
	return imageURL
}

func htmlUnescape(s string) string {
	r := strings.NewReplacer(
		"&amp;", "&",
		"&quot;", `"`,
		"&#39;", "'",
		"&lt;", "<",
		"&gt;", ">",
	)
	return r.Replace(s)
}
