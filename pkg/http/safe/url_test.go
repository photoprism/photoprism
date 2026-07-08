package safe

import (
	"errors"
	"testing"
)

func TestURL(t *testing.T) {
	t.Run("AcceptHTTP", func(t *testing.T) {
		u, err := URL("http://localhost:2342/api")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u == nil || u.Host != "localhost:2342" {
			t.Fatalf("unexpected parsed URL: %#v", u)
		}
	})
	t.Run("AcceptHTTPS", func(t *testing.T) {
		u, err := URL("https://example.com/v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u == nil || u.Host != "example.com" {
			t.Fatalf("unexpected parsed URL: %#v", u)
		}
	})
	t.Run("RejectMissingHost", func(t *testing.T) {
		if _, err := URL("https:///v1"); !errors.Is(err, ErrURLHostRequired) {
			t.Fatalf("expected ErrURLHostRequired, got %v", err)
		}
	})
	t.Run("RejectUnsupportedScheme", func(t *testing.T) {
		if _, err := URL("file:///tmp/payload.json"); !errors.Is(err, ErrSchemeNotAllowed) {
			t.Fatalf("expected ErrSchemeNotAllowed, got %v", err)
		}
	})
}
