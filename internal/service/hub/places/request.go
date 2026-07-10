package places

import (
	"crypto/sha1" //nolint:gosec // required for upstream signature scheme
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/safe"
)

// sharedTransport is a connection-pooling transport reused across all geocoding
// requests. The idle-connection pool lives in the transport, not the client, so
// a large index keeps warm connections instead of a fresh TLS handshake per
// lookup — while each request still uses its own client for a request-scoped
// timeout (see GetRequest).
var sharedTransport = newTransport()

// newTransport returns an http.Transport tuned for many short geocoding
// requests. It keeps a larger idle-connection pool than the stdlib default to
// reduce TLS-handshake churn, while leaving IdleConnTimeout below the backend's
// so the client closes idle connections first and avoids the reuse race.
func newTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 64
	t.MaxIdleConnsPerHost = 16
	t.IdleConnTimeout = 90 * time.Second
	return t
}

// serviceHost returns the host of a request URL for UI-visible logs. It drops
// the query so request coordinates never reach those logs; if the URL cannot be
// parsed into a host, it still strips any query string rather than falling back
// to the raw value.
func serviceHost(reqUrl string) string {
	if u, err := url.Parse(reqUrl); err == nil && u.Host != "" {
		return u.Host
	}
	if i := strings.IndexByte(reqUrl, '?'); i >= 0 {
		return reqUrl[:i]
	}
	return reqUrl
}

// safeError returns an error message safe for UI-visible logs. A failed request
// yields a *url.Error whose Error() embeds the full request URL — including the
// geocoding coordinates — so for that type only the underlying transport error
// is returned; other errors are returned unchanged.
func safeError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		return urlErr.Err.Error()
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

// GetRequest fetches data from the specified geocoding service URL. Because
// these are idempotent GET requests, it retries transient connection-level
// failures up to Retries times with a linear backoff before giving up.
func GetRequest(reqUrl string, locale string) (r *http.Response, err error) {
	var req *http.Request

	// Full request URL at trace level only; trace is developer-only and not
	// surfaced in the UI, unlike the host-only error logs.
	log.Tracef("places: sending request to %s", reqUrl)

	if _, parseErr := safe.URL(reqUrl); parseErr != nil {
		return r, fmt.Errorf("places: unsupported request URL scheme")
	}

	// Create GET request instance.
	req, err = http.NewRequest(http.MethodGet, reqUrl, nil)

	// Ok?
	if err != nil {
		log.Errorf("places: request to %s failed (%s)", serviceHost(reqUrl), safeError(err))
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

	// Per-request client sharing the pooled transport. The client stays per
	// request because Client.Timeout covers reading the response body and keeps
	// running after Do returns, so the deadline must be scoped to this request;
	// the connection pool lives in sharedTransport, so lookups still reuse warm
	// connections.
	client := &http.Client{Timeout: 60 * time.Second, Transport: sharedTransport}

	// Retry transient connection-level failures on this idempotent GET with a
	// linear backoff.
	for i := 0; i < Retries; i++ {
		// #nosec G704 reqUrl is parsed and scheme-validated above.
		if r, err = client.Do(req); err == nil {
			return r, nil
		}

		// Full URL and raw error at trace level for developer troubleshooting.
		log.Tracef("places: request to %s failed (%s)", reqUrl, err)

		// Back off before the next attempt, skipping the wait after the last try.
		if RetryDelay > 0 && i < Retries-1 {
			time.Sleep(RetryDelay * time.Duration(i+1))
		}
	}

	return r, err
}
