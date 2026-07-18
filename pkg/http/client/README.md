## PhotoPrism — pkg/http/client

**Last Updated:** July 18, 2026

### Overview

`pkg/http/client` provides bounded, opt-in retry helpers for outgoing HTTP requests. Its `Do` wrapper retries only the HTTP status codes a caller explicitly lists (typically `429 Too Many Requests`) using jittered exponential backoff, honors a `Retry-After` response header, and stays within the caller's `context` deadline. It is self-contained on the Go standard library so any package — including `pkg/*` — can use it without importing `internal/*`.

#### Goals

- Retry transient, retryable HTTP status codes (e.g. `429`) with bounded exponential backoff and jitter, instead of failing on the first refusal.
- Honor a `Retry-After` header when present, capped so a hostile or misconfigured endpoint cannot stall the caller.
- Keep total elapsed time within the caller's `context` deadline.
- Replay buffered request bodies safely across attempts and follow the `net/http` error convention (nil response on error, with any interim body already drained and closed).

#### Non-Goals

- Connection-level retries (`io.EOF`, `ECONNRESET`, `tls: EOF`) for idempotent `GET`/`HEAD`, and a shared pooled `*http.Client` factory. These are tracked separately in [#5732](https://github.com/photoprism/photoprism/issues/5732); `Do` returns connection-level errors as-is.
- Deciding request idempotency. `Do` retries whatever status codes the caller opts into, so the caller owns replay-safety (only send retryable statuses for requests that are safe to repeat).

### Package Layout (Code Map)

- Retry wrapper: `retry.go` — `Do`, `RetryPolicy`, and the `backoff` / `jitter` / `sleep` / `drainAndClose` helpers.
- `Retry-After` parsing: `retry_after.go` — `ParseRetryAfter` (both RFC 7231 forms: delta-seconds and HTTP-date).
- Package doc & license header: `client.go`.
- Tests: `retry_test.go`, `retry_after_test.go`.

### Usage

`Do` takes a `newReq` factory (called once per attempt so a buffered body replays cleanly) and a `RetryPolicy`:

```go
data := []byte(`{"prompt":"…"}`)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

httpClient := &http.Client{Timeout: 10 * time.Minute}

newReq := func() (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	header.SetContentType(req, header.ContentTypeJson)
	return req, nil
}

resp, err := client.Do(ctx, httpClient, newReq, client.RetryPolicy{
	MaxRetries:      2,                                   // extra attempts after the first
	BaseDelay:       500 * time.Millisecond,              // doubled each attempt
	MaxDelay:        8 * time.Second,                     // per-attempt cap (also caps Retry-After)
	RetryStatuses:   []int{http.StatusTooManyRequests},   // opt-in: 429 only
	HonorRetryAfter: true,
})
if err != nil {
	return err // response is nil; any interim body is already closed
}
defer resp.Body.Close()
// resp is the final (possibly still-4xx) response — inspect resp.StatusCode.
```

Behavior notes:

- **Opt-in only.** A status not in `RetryStatuses` (including other `>= 400`) is returned immediately as the final response.
- **Retry-After.** When present and larger than the computed backoff, it is used instead — but capped at `MaxDelay`, so a longer requested pause is retried sooner and may fail through rather than blocking the caller.
- **Deadline vs. status.** If the budget runs out **between** attempts, `Do` returns the last response with a nil error (the item surfaces as its real status, e.g. `429`). If the deadline fires **while a request is in flight**, `http.Client.Do` returns a `context.DeadlineExceeded` that `Do` passes through as `(nil, err)`.
- **Error contract.** On a non-nil error the response is nil and any interim body has been drained and closed, so the caller never closes a body on the error path. On a nil error the caller owns closing `resp.Body`.

The vision service client (`internal/ai/vision/api_client.go`) is the first consumer: it opts into `429` retries for its read-only inference `POST`s.

### Test Guidelines

- Drive retries with an `httptest` server whose handler counts attempts and returns `429` before `200`; assert the final status and the request count.
- For deterministic control over the backoff wait (e.g. context cancellation) without network timing, use a stub `http.RoundTripper` that returns a canned response, so cancellation affects only the wait.
- Focused run: `go test ./pkg/http/client -count=1`.
