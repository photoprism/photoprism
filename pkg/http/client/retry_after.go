package client

import (
	"net/http"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// ParseRetryAfter returns the delay requested by a response Retry-After header.
// It accepts both RFC 7231 forms, delta-seconds and an HTTP-date, and reports
// false only when the header is absent, negative, or malformed. A past or zero
// HTTP-date yields (0, true) so the caller retries immediately.
func ParseRetryAfter(resp *http.Response) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}

	value := resp.Header.Get(header.RetryAfter)

	if value == "" {
		return 0, false
	}

	// Delta-seconds form, e.g. "Retry-After: 5".
	if secs, err := strconv.Atoi(value); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}

	// HTTP-date form, e.g. "Retry-After: Wed, 21 Oct 2026 07:28:00 GMT".
	if t, err := http.ParseTime(value); err == nil {
		if d := time.Until(t); d > 0 {
			return d, true
		}
		return 0, true
	}

	return 0, false
}
