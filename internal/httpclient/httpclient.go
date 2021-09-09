package httpclient

import (
	"bytes"
	"io"
	"net/http"

	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

func Client(debug bool) *http.Client {
	if debug {
		return &http.Client{
			Transport: LoggingRoundTripper{http.DefaultTransport},
		}
	}
	return http.DefaultClient
}

// This type implements the http.RoundTripper interface
type LoggingRoundTripper struct {
	proxy http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	// Do "before sending requests" actions here.
	log.Debugf("Sending request to %v\n", req.URL.String())

	// Copy body into buffer for logging
	//bu := new(bytes.Buffer)
	//_, err := io.Copy(bu, req.Body)
	//if err != nil {
	//	log.Errorf("Error: %v", err)
	//}
	//log.Debugf("Request Header: %s\n", req.Header)
	//log.Debugf("Request Body: %s\n", bu.String())
	//req.Body = io.NopCloser(bu)

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
		log.Debugf("Reponse Body: %s\n", buf.String())
		res.Body = io.NopCloser(buf)
	}
	return
}
