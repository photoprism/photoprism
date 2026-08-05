package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// trackedBody is a response body that records whether it was closed, so a test
// can assert Do drained and closed an interim response.
type trackedBody struct {
	closed bool
}

func (b *trackedBody) Read(p []byte) (int, error) { return 0, io.EOF }
func (b *trackedBody) Close() error               { b.closed = true; return nil }

// stubTransport returns a canned status and body without touching the network,
// so a canceled context affects only the backoff wait, not the request itself.
type stubTransport struct {
	status int
	body   *trackedBody
}

func (s *stubTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: s.status, Header: http.Header{}, Body: s.body}, nil
}

// newReqFactory returns a request builder that replays a small POST body, so a
// retried attempt sends the same payload the first one did.
func newReqFactory(ctx context.Context, url string) func() (*http.Request, error) {
	return func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(`{"ping":true}`)))
	}
}

func fastPolicy(retries int) RetryPolicy {
	return RetryPolicy{
		MaxRetries:      retries,
		BaseDelay:       time.Millisecond,
		MaxDelay:        5 * time.Millisecond,
		RetryStatuses:   []int{http.StatusTooManyRequests},
		HonorRetryAfter: true,
	}
}

func TestDo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := Do(context.Background(), server.Client(), newReqFactory(context.Background(), server.URL), fastPolicy(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		drainAndClose(resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("expected 1 attempt, got %d", got)
		}
	})
	t.Run("RetryThenSuccess", func(t *testing.T) {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&calls, 1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := Do(context.Background(), server.Client(), newReqFactory(context.Background(), server.URL), fastPolicy(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		drainAndClose(resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if got := atomic.LoadInt32(&calls); got != 2 {
			t.Fatalf("expected 2 attempts, got %d", got)
		}
	})
	t.Run("NonRetryableStatus", func(t *testing.T) {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		resp, err := Do(context.Background(), server.Client(), newReqFactory(context.Background(), server.URL), fastPolicy(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		drainAndClose(resp)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("expected 1 attempt, got %d", got)
		}
	})
	t.Run("Exhausted", func(t *testing.T) {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		resp, err := Do(context.Background(), server.Client(), newReqFactory(context.Background(), server.URL), fastPolicy(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		drainAndClose(resp)
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", resp.StatusCode)
		}
		if got := atomic.LoadInt32(&calls); got != 3 {
			t.Fatalf("expected 3 attempts, got %d", got)
		}
	})
	t.Run("RetryAfterHonored", func(t *testing.T) {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&calls, 1) == 1 {
				w.Header().Set(header.RetryAfter, "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := Do(context.Background(), server.Client(), newReqFactory(context.Background(), server.URL), fastPolicy(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		drainAndClose(resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if got := atomic.LoadInt32(&calls); got != 2 {
			t.Fatalf("expected 2 attempts, got %d", got)
		}
	})
	t.Run("DeadlineStopsRetry", func(t *testing.T) {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		// A deadline shorter than the backoff must prevent the wait.
		p := fastPolicy(5)
		p.BaseDelay = time.Hour
		p.MaxDelay = time.Hour
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// A backoff longer than the remaining budget stops retrying and returns
		// the last response so the caller can treat the 429 as terminal.
		resp, err := Do(ctx, server.Client(), newReqFactory(ctx, server.URL), p)
		if err != nil {
			t.Fatalf("expected nil error on the budget-exhausted path, got %v", err)
		}
		drainAndClose(resp)
		if resp == nil || resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("expected last 429 response, got %v", resp)
		}
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("expected 1 attempt before deadline, got %d", got)
		}
	})
	t.Run("ContextCanceledDrainsAndErrors", func(t *testing.T) {
		// The stub transport ignores the context, so cancellation only interrupts
		// the backoff wait — exercising the drain-and-error path in Do.
		body := &trackedBody{}
		c := &http.Client{Transport: &stubTransport{status: http.StatusTooManyRequests, body: body}}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		p := RetryPolicy{MaxRetries: 3, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second, RetryStatuses: []int{http.StatusTooManyRequests}}
		newReq := func() (*http.Request, error) {
			return http.NewRequest(http.MethodPost, "http://example.invalid", bytes.NewReader([]byte("{}")))
		}

		resp, err := Do(ctx, c, newReq, p)
		if err == nil {
			t.Fatal("expected a context error")
		}
		if resp != nil {
			t.Fatalf("expected nil response on the error path, got %v", resp)
		}
		if !body.closed {
			t.Fatal("expected the interim response body to be drained and closed")
		}
	})
}

func TestRetryPolicyShouldRetry(t *testing.T) {
	p := RetryPolicy{RetryStatuses: []int{http.StatusTooManyRequests, http.StatusServiceUnavailable}}
	t.Run("Match", func(t *testing.T) {
		if !p.shouldRetry(http.StatusTooManyRequests) {
			t.Fatal("expected 429 to be retryable")
		}
	})
	t.Run("NoMatch", func(t *testing.T) {
		if p.shouldRetry(http.StatusBadRequest) {
			t.Fatal("expected 400 not to be retryable")
		}
	})
}

func TestJitter(t *testing.T) {
	t.Run("WithinBounds", func(t *testing.T) {
		base := 100 * time.Millisecond
		for i := 0; i < 100; i++ {
			d := jitter(base)
			if d < base*3/4 || d > base*5/4 {
				t.Fatalf("jitter %v outside +/-25%% of %v", d, base)
			}
		}
	})
	t.Run("NonPositive", func(t *testing.T) {
		if d := jitter(0); d != 0 {
			t.Fatalf("expected 0, got %v", d)
		}
	})
}

func TestSleep(t *testing.T) {
	t.Run("NonPositiveNoWait", func(t *testing.T) {
		if err := sleep(context.Background(), 0); err != nil {
			t.Fatalf("expected nil for zero delay, got %v", err)
		}
	})
	t.Run("CompletesWait", func(t *testing.T) {
		if err := sleep(context.Background(), time.Millisecond); err != nil {
			t.Fatalf("expected nil after completing the wait, got %v", err)
		}
	})
	t.Run("CancelledBeforeWait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := sleep(ctx, 0); err == nil {
			t.Fatal("expected ctx error for an already-canceled context")
		}
	})
	t.Run("CancelledDuringWait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := sleep(ctx, time.Hour); err == nil {
			t.Fatal("expected ctx error when canceled during the wait")
		}
	})
}

func TestBackoff(t *testing.T) {
	p := RetryPolicy{BaseDelay: 100 * time.Millisecond, MaxDelay: 400 * time.Millisecond}
	resp := &http.Response{Header: http.Header{}}
	t.Run("GrowsAndCaps", func(t *testing.T) {
		// With +/-25% jitter the cap can be exceeded by at most 25%.
		for attempt := 0; attempt < 6; attempt++ {
			d := p.backoff(attempt, resp)
			if d <= 0 {
				t.Fatalf("attempt %d: expected positive delay, got %v", attempt, d)
			}
			if d > p.MaxDelay*5/4 {
				t.Fatalf("attempt %d: delay %v exceeds capped jitter bound", attempt, d)
			}
		}
	})
	t.Run("RetryAfterRaisesDelay", func(t *testing.T) {
		p := RetryPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Second, HonorRetryAfter: true}
		resp := &http.Response{Header: http.Header{}}
		resp.Header.Set(header.RetryAfter, "1")
		if d := p.backoff(0, resp); d != time.Second {
			t.Fatalf("expected Retry-After to raise delay to 1s (capped), got %v", d)
		}
	})
}
