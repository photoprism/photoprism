package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

func Client(debug bool) *http.Client {
	if debug {
		return &http.Client{
			Transport: LoggingRoundTripper{http.DefaultTransport},
			Timeout:   time.Second * 10,
		}
	}
	cl := http.DefaultClient
	cl.Timeout = time.Second * 10
	return cl
}

// This type implements the http.RoundTripper interface
type LoggingRoundTripper struct {
	proxy http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	// Do "before sending requests" actions here.
	log.Debugf("Sending request to %v\n", req.URL.String())

	// Send the request, get the response (or the error)
	res, e = lrt.proxy.RoundTrip(req)

	// Handle the result.
	if e != nil {
		log.Errorf("Error: %v", e)
	} else {
		log.Debugf("Received %v response\n", res.Status)

		// Copy body into buffer for logging
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, res.Body)
		if err != nil {
			log.Errorf("Error: %v", err)
		}
		log.Debugf("Header: %s\n", res.Header)
		log.Debugf("Reponse Body: %s\n", buf.String())
		res.Body = io.NopCloser(buf)
	}
	return
}
