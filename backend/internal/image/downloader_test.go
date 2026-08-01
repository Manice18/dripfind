package image

import "testing"

func TestDownload_BlocksPrivateAndLinkLocal(t *testing.T) {
	d := NewDownloader()

	urls := []string{
		"http://127.0.0.1:8080/health",
		"http://169.254.169.254/latest/meta-data/",
	}
	for _, u := range urls {
		_, err := d.Download(u)
		if err == nil {
			t.Fatalf("%s: expected error, got nil", u)
		}
	}
}
