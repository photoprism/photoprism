package places

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRequest(t *testing.T) {
	t.Run("RejectUnsupportedScheme", func(t *testing.T) {
		if _, err := GetRequest("file:///tmp/location.json", "en"); err == nil {
			t.Fatal("expected error for unsupported scheme")
		}
	})
	t.Run("Success", func(t *testing.T) {
		prevRetries := Retries
		prevDelay := RetryDelay
		prevAgent := UserAgent
		prevKey := Key
		prevSecret := Secret
		defer func() {
			Retries = prevRetries
			RetryDelay = prevDelay
			UserAgent = prevAgent
			Key = prevKey
			Secret = prevSecret
		}()

		Retries = 1
		RetryDelay = 0
		UserAgent = "PhotoPrism/TestSuite"
		Key = ""
		Secret = ""

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Accept-Language"); got != "de" {
				t.Fatalf("expected locale header 'de', got %q", got)
			}

			if got := r.Header.Get("User-Agent"); got != UserAgent {
				t.Fatalf("expected user agent %q, got %q", UserAgent, got)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := GetRequest(server.URL+"/v1/location/test", "de")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("expected response")
		}
		_ = resp.Body.Close()
	})
	t.Run("InvalidURL", func(t *testing.T) {
		prevRetries := Retries
		prevDelay := RetryDelay
		defer func() {
			Retries = prevRetries
			RetryDelay = prevDelay
		}()

		Retries = 1
		RetryDelay = 10 * time.Millisecond
		if _, err := GetRequest("://invalid", "en"); err == nil {
			t.Fatal("expected URL parse error")
		}
	})
}
