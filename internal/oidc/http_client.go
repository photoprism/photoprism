package oidc

import (
	"bytes"
	"io"
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

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	// Do "before sending requests" actions here.
	log.Debugf("sending request to %s", req.URL.String())

	// Send the request, get the response (or the error)
	res, e = lrt.proxy.RoundTrip(req)

	// Handle the result.
	if e != nil {
		log.Errorf("http error: %s", e)
	} else {
		log.Debugf("http response: %s", res.Status)

		// Copy body into buffer for logging
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, res.Body)
		if err != nil {
			log.Errorf("http buffer error: %s", err)
		}
		// log.Debugf("Header: %s\n", res.Header)
		// log.Debugf("Reponse Body: %s\n", buf.String())
		res.Body = io.NopCloser(buf)
	}
	return
}
