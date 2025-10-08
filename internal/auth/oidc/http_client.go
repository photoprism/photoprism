package oidc

import (
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/event"
)

// HttpClient returns an HTTP client tailored for OIDC requests. When debug is true, it wraps the
// default transport with a LoggingRoundTripper and keeps a 30s timeout.
func HttpClient(debug bool) *http.Client {
	if debug {
		return &http.Client{
			Transport: LoggingRoundTripper{http.DefaultTransport},
			Timeout:   time.Second * 30,
		}
	}

	return &http.Client{Timeout: 30 * time.Second}
}

// LoggingRoundTripper wraps an http.RoundTripper and emits audit logs for OIDC requests.
type LoggingRoundTripper struct {
	proxy http.RoundTripper
}

// RoundTrip logs the request method, URL and error, if any.
func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	// Perform HTTP request.
	res, err = lrt.proxy.RoundTrip(req)

	// Log the request method, URL and error, if any.
	if err != nil {
		event.AuditErr([]string{"oidc", "provider", "request", "%s %s", "%s"}, req.Method, req.URL.String(), err)
	} else {
		event.AuditDebug([]string{"oidc", "provider", "request", "%s %s", "%s"}, req.Method, req.URL.String(), res.Status)
	}

	return res, err
}
