package places

import (
	"crypto/sha1" //nolint:gosec // required for upstream signature scheme
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/safe"
)

// httpClient is a shared, connection-pooling HTTP client reused across all
// geocoding requests. Reusing one client keeps a warm idle-connection pool, so a
// large index does not open a fresh TLS handshake for every lookup.
var httpClient = &http.Client{
	Timeout:   60 * time.Second,
	Transport: newTransport(),
}

// newTransport returns an http.Transport tuned for many short geocoding
// requests: a larger idle-connection pool than the stdlib default to reduce
// TLS-handshake churn, while keeping the client's idle timeout below the
// backend's so the client closes idle connections first and avoids the
// connection-reuse race.
func newTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 64
	t.MaxIdleConnsPerHost = 16
	t.IdleConnTimeout = 90 * time.Second
	return t
}

// serviceHost returns the host of a request URL for logging. It drops the query
// string so request coordinates never reach the logs, and falls back to the raw
// value if the URL cannot be parsed.
func serviceHost(reqUrl string) string {
	if u, err := url.Parse(reqUrl); err == nil && u.Host != "" {
		return u.Host
	}
	return reqUrl
}

// GetRequest fetches data from the specified service URL. Because these are
// idempotent GET requests, it retries transient connection-level failures up to
// Retries times with a linear backoff before giving up.
func GetRequest(reqUrl string, locale string) (r *http.Response, err error) {
	var req *http.Request

	// Log request host (without query parameters).
	log.Tracef("places: sending request to %s", serviceHost(reqUrl))

	if _, parseErr := safe.URL(reqUrl); parseErr != nil {
		return r, fmt.Errorf("places: unsupported request URL scheme")
	}

	// Create GET request instance.
	req, err = http.NewRequest(http.MethodGet, reqUrl, nil)

	// Ok?
	if err != nil {
		log.Errorf("places: %s", err.Error())
		return r, err
	}

	// Set user agent.
	if UserAgent != "" {
		req.Header.Set(header.UserAgent, UserAgent)
	} else {
		req.Header.Set(header.UserAgent, "PhotoPrism/Test")
	}

	// Set requested result locale.
	if locale != "" {
		req.Header.Set(header.AcceptLanguage, locale)
	}

	// Add API key?
	if Key != "" {
		req.Header.Set("X-Key", Key)
		req.Header.Set("X-Signature", fmt.Sprintf("%x", sha1.Sum([]byte(Key+reqUrl+Secret)))) //nolint:gosec // upstream expects SHA1
	}

	// Perform the request, retrying transient connection failures on this
	// idempotent GET with a linear backoff between attempts.
	for i := 0; i < Retries; i++ {
		// #nosec G704 reqUrl is parsed and scheme-validated above.
		if r, err = httpClient.Do(req); err == nil {
			return r, nil
		}

		// Wait before trying again?
		if RetryDelay > 0 {
			time.Sleep(RetryDelay * time.Duration(i+1))
		}
	}

	return r, err
}
