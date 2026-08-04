package search

import "testing"

func TestBewakoofProductURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"mens-white-grey-striped-shirt", "https://www.bewakoof.com/p/mens-white-grey-striped-shirt"},
		{"/mens-white-grey-striped-shirt", "https://www.bewakoof.com/p/mens-white-grey-striped-shirt"},
		{"/p/mens-white-grey-striped-shirt", "https://www.bewakoof.com/p/mens-white-grey-striped-shirt"},
		{"p/mens-white-grey-striped-shirt", "https://www.bewakoof.com/p/mens-white-grey-striped-shirt"},
		{
			"https://www.bewakoof.com/mens-white-grey-striped-shirt",
			"https://www.bewakoof.com/p/mens-white-grey-striped-shirt",
		},
		{
			"https://www.bewakoof.com/p/mens-black-slim-fit-t-shirt9?src=myorder",
			"https://www.bewakoof.com/p/mens-black-slim-fit-t-shirt9?src=myorder",
		},
	}
	for _, tc := range cases {
		if got := bewakoofProductURL(tc.in); got != tc.want {
			t.Fatalf("bewakoofProductURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
