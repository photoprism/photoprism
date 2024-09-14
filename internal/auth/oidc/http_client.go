package oidc

import (
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/event"
)

// HttpClient represents a client that makes HTTP requests.
//
// NOTE: Timeout specifies a time limit for requests made by
// this Client. The timeout includes connection time, any
// redirects, and reading the response body. The timer remains
// running after Get, Head, Post, or Do return and will
// interrupt reading of the Response.Body.
func HttpClient(debug bool) *http.Client {
	if debug {
		return &http.Client{
			Transport: LoggingRoundTripper{http.DefaultTransport},
			Timeout:   time.Second * 30,
		}
	}

	return &http.Client{Timeout: 30 * time.Second}
}

// LoggingRoundTripper specifies the http.RoundTripper interface.
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
