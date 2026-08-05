package client

import (
	"net/http"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestParseRetryAfter(t *testing.T) {
	newResp := func(value string) *http.Response {
		h := http.Header{}
		if value != "" {
			h.Set(header.RetryAfter, value)
		}
		return &http.Response{Header: h}
	}
	t.Run("Seconds", func(t *testing.T) {
		d, ok := ParseRetryAfter(newResp("5"))
		if !ok || d != 5*time.Second {
			t.Fatalf("expected 5s ok, got %v %v", d, ok)
		}
	})
	t.Run("ZeroSeconds", func(t *testing.T) {
		d, ok := ParseRetryAfter(newResp("0"))
		if !ok || d != 0 {
			t.Fatalf("expected 0s ok, got %v %v", d, ok)
		}
	})
	t.Run("NegativeSeconds", func(t *testing.T) {
		if d, ok := ParseRetryAfter(newResp("-3")); ok {
			t.Fatalf("expected not ok for negative, got %v %v", d, ok)
		}
	})
	t.Run("HTTPDateFuture", func(t *testing.T) {
		future := time.Now().UTC().Add(30 * time.Second).Format(http.TimeFormat)
		d, ok := ParseRetryAfter(newResp(future))
		if !ok || d <= 0 || d > 31*time.Second {
			t.Fatalf("expected positive sub-31s duration, got %v %v", d, ok)
		}
	})
	t.Run("HTTPDatePast", func(t *testing.T) {
		past := time.Now().UTC().Add(-30 * time.Second).Format(http.TimeFormat)
		d, ok := ParseRetryAfter(newResp(past))
		if !ok || d != 0 {
			t.Fatalf("expected 0 ok for past date, got %v %v", d, ok)
		}
	})
	t.Run("Missing", func(t *testing.T) {
		if _, ok := ParseRetryAfter(newResp("")); ok {
			t.Fatal("expected not ok for missing header")
		}
	})
	t.Run("Malformed", func(t *testing.T) {
		if _, ok := ParseRetryAfter(newResp("soon")); ok {
			t.Fatal("expected not ok for malformed header")
		}
	})
	t.Run("NilResponse", func(t *testing.T) {
		if _, ok := ParseRetryAfter(nil); ok {
			t.Fatal("expected not ok for nil response")
		}
	})
}
