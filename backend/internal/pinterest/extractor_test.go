package pinterest

import "testing"

func TestValidateURL_PinterestOnly(t *testing.T) {
	ok := []string{
		"https://www.pinterest.com/pin/123456/",
		"https://pinterest.com/pin/123456/",
		"https://in.pinterest.com/pin/123456/",
		"https://pinterest.co.uk/pin/123456/",
		"https://pin.it/abc123",
		"https://i.pinimg.com/736x/ab/cd/ef.jpg",
	}
	for _, u := range ok {
		if err := ValidateURL(u); err != nil {
			t.Fatalf("%s: unexpected error: %v", u, err)
		}
	}

	bad := []string{
		"https://instagram.com/p/abc",
		"https://www.tiktok.com/@user/video/1",
		"https://images.unsplash.com/photo.jpg",
		"https://example.com/outfit.png",
		"https://notpinterest.com/pin/1",
		"not-a-url",
	}
	for _, u := range bad {
		if err := ValidateURL(u); err == nil {
			t.Fatalf("%s: expected error", u)
		}
	}
}
