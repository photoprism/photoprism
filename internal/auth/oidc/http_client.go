package oidc

import (
	"net/http"
	"time"
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
			Timeout:   time.Second * 20,
		}
	}

	return &http.Client{Timeout: 20 * time.Second}
}

// LoggingRoundTripper specifies the http.RoundTripper interface.
type LoggingRoundTripper struct {
	proxy http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	log.Tracef("oidc: %s %s", req.Method, req.URL.String())

	// Send request.
	res, err = lrt.proxy.RoundTrip(req)

	// Log error, if any.
	if err != nil {
		log.Debugf("oidc: request to %s has failed (%s)", req.URL.String(), err)
	}

	return res, err
}
