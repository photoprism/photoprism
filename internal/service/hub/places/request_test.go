package places

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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
	t.Run("RetryOnConnectionError", func(t *testing.T) {
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

		Retries = 3
		RetryDelay = 0
		UserAgent = "PhotoPrism/TestSuite"
		Key = ""
		Secret = ""

		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Drop the connection on the first attempt to simulate a transient
			// connection-level failure, then answer normally on the retry.
			if atomic.AddInt32(&attempts, 1) == 1 {
				if hj, ok := w.(http.Hijacker); ok {
					if conn, _, hjErr := hj.Hijack(); hjErr == nil {
						_ = conn.Close()
						return
					}
				}
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := GetRequest(server.URL+"/v1/reverse", "en")
		if err != nil {
			t.Fatalf("expected success after retry, got %v", err)
		}
		if resp == nil {
			t.Fatal("expected response")
		}
		_ = resp.Body.Close()
		if got := atomic.LoadInt32(&attempts); got < 2 {
			t.Fatalf("expected at least 2 attempts (a retry), got %d", got)
		}
	})
}

func TestServiceHost(t *testing.T) {
	t.Run("WithQuery", func(t *testing.T) {
		if got := serviceHost("https://places.photoprism.app/v1/reverse?lat=52.520000&lng=13.405000"); got != "places.photoprism.app" {
			t.Fatalf("expected host without query, got %q", got)
		}
	})
	t.Run("PathOnly", func(t *testing.T) {
		if got := serviceHost("https://places.photoprism.xyz/v1/location/1e95998417cc"); got != "places.photoprism.xyz" {
			t.Fatalf("expected host, got %q", got)
		}
	})
	t.Run("Unparseable", func(t *testing.T) {
		if got := serviceHost("://nope"); got != "://nope" {
			t.Fatalf("expected raw fallback, got %q", got)
		}
	})
}

func TestNewTransport(t *testing.T) {
	tr := newTransport()
	if tr == nil {
		t.Fatal("expected transport")
	}
	if tr.MaxIdleConnsPerHost != 16 {
		t.Fatalf("expected MaxIdleConnsPerHost 16, got %d", tr.MaxIdleConnsPerHost)
	}
	if tr.MaxIdleConns != 64 {
		t.Fatalf("expected MaxIdleConns 64, got %d", tr.MaxIdleConns)
	}
	if tr.IdleConnTimeout != 90*time.Second {
		t.Fatalf("expected IdleConnTimeout 90s, got %s", tr.IdleConnTimeout)
	}
}
