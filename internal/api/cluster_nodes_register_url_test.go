package api

import "testing"

func TestValidateAdvertiseURL(t *testing.T) {
	cases := []struct {
		u  string
		ok bool
	}{
		{"https://example.com", true},
		{"http://example.com", false},
		{"http://localhost:2342", true},
		{"https://127.0.0.1", true},
		{"ftp://example.com", false},
		{"https://", false},
		{"", false},
	}
	for _, c := range cases {
		if got := validateAdvertiseURL(c.u); got != c.ok {
			t.Fatalf("validateAdvertiseURL(%q) = %v, want %v", c.u, got, c.ok)
		}
	}
}

func TestValidateSiteURL(t *testing.T) {
	cases := []struct {
		u  string
		ok bool
	}{
		{"https://photos.example.com", true},
		{"http://photos.example.com", false},
		{"http://127.0.0.1:2342", true},
		{"mailto:me@example.com", false},
		{"://bad", false},
	}
	for _, c := range cases {
		if got := validateSiteURL(c.u); got != c.ok {
			t.Fatalf("validateSiteURL(%q) = %v, want %v", c.u, got, c.ok)
		}
	}
}
