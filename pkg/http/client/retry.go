package client

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"time"
)

// RetryPolicy configures bounded status-code retries with exponential backoff.
// A zero MaxRetries disables retrying, so the request is attempted exactly once.
type RetryPolicy struct {
	MaxRetries      int           // Number of extra attempts after the first request.
	BaseDelay       time.Duration // Backoff before the first retry, doubled each attempt.
	MaxDelay        time.Duration // Upper bound for a single backoff wait.
	RetryStatuses   []int         // Response status codes that trigger a retry (opt-in).
	HonorRetryAfter bool          // Wait for a Retry-After header when present.
}

// shouldRetry reports whether a response status code is configured as retryable.
func (p RetryPolicy) shouldRetry(status int) bool {
	return slices.Contains(p.RetryStatuses, status)
}

// Do runs newReq through client.Do with bounded exponential backoff for the
// policy's retryable status codes. It rebuilds the request via newReq on every
// attempt so a buffered body is replayed safely, and drains each non-final
// response body before waiting so the connection returns to the pool. Only the
// caller-listed status codes are retried; connection-level errors are returned
// as-is because the caller owns request idempotency.
//
// The returned response follows the net/http convention: on a non-nil error the
// response is nil and any interim body has been drained and closed, so the
// caller never has to close a body on the error path. When the status is
// terminal or the retries/deadline budget is exhausted, Do returns the last
// response with a nil error and the caller owns closing its body.
func Do(ctx context.Context, client *http.Client, newReq func() (*http.Request, error), p RetryPolicy) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		var req *http.Request

		if req, err = newReq(); err != nil {
			return nil, err
		}

		if resp, err = client.Do(req); err != nil {
			drainAndClose(resp)
			return nil, err
		}

		// Stop when the status is terminal or no retries remain.
		if attempt >= p.MaxRetries || !p.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		// Return the last response as terminal when the backoff would outlast the
		// deadline budget, leaving its body open for the caller to inspect.
		wait := p.backoff(attempt, resp)

		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= wait {
			return resp, nil
		}

		// A retry is due: free the connection during the wait.
		drainAndClose(resp)

		// Wait, honoring cancellation. A canceled context is terminal.
		if waitErr := sleep(ctx, wait); waitErr != nil {
			return nil, waitErr
		}
	}
}

// backoff returns the wait before the next attempt: an exponential base delay
// with jitter, raised to a present Retry-After value when honored, and capped at
// MaxDelay so a hostile or misconfigured endpoint cannot stall the caller.
func (p RetryPolicy) backoff(attempt int, resp *http.Response) time.Duration {
	delay := p.BaseDelay << attempt

	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	delay = jitter(delay)

	if p.HonorRetryAfter {
		if ra, ok := ParseRetryAfter(resp); ok && ra > delay {
			delay = ra
		}
	}

	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	return delay
}

// jitter applies up to +/-25% jitter to a duration to avoid retry stampedes.
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}

	//nolint:gosec // math/rand is sufficient for backoff jitter, no security relevance.
	n := time.Duration(float64(d) * (0.75 + rand.Float64()*0.5))

	if n <= 0 {
		return d
	}

	return n
}

// sleep waits for d and returns ctx.Err() if the context is canceled first, or
// nil once the wait completes. A non-positive delay returns immediately, still
// surfacing a cancellation that already happened.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// drainAndClose discards and closes a response body so the underlying
// connection can be returned to the pool for reuse.
func drainAndClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
