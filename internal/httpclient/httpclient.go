package httpclient

import (
	"bytes"
	"io"
	"net/http"

	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

var httpclient *http.Client

func Client(debug bool) *http.Client {
	if httpclient != nil {
		return httpclient
	}
	if debug {
		httpclient = &http.Client{
			Transport: LoggingRoundTripper{http.DefaultTransport},
		}
	} else {
		httpclient = http.DefaultClient
	}
	return httpclient
}

// This type implements the http.RoundTripper interface
type LoggingRoundTripper struct {
	proxy http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	// Do "before sending requests" actions here.
	log.Debugf("Sending request to %v\n", req.URL)

	// Send the request, get the response (or the error)
	res, e = lrt.proxy.RoundTrip(req)

	// Handle the result.
	if e != nil {
		log.Errorf("Error: %v", e)
	} else {
		log.Debugf("Received %v response\n", res.Status)

		//buf := bytes.NewBuffer(make([]byte, 0, 1024))
		//r := io.TeeReader(res.Body, buf)
		//all, err := io.ReadAll(r)
		//log.Debugf("Body: %s\nError: %s\n", all, err)
		//res.Body = io.NopCloser(buf)

		//pipeReader, pipeWriter := io.Pipe()
		//r := io.TeeReader(res.Body, pipeWriter)
		//go func() {
		//	all, err := io.ReadAll(r)
		//	pipeWriter.CloseWithError(err)
		//	log.Debugf("Body\n%s\n%s", all, err)
		//}()
		//res.Body = pipeReader

		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, res.Body)
		if err != nil {
			log.Errorf("Error: %v", err)
		}
		log.Debugf("Reponse Body: %s\n", buf.String())
		res.Body = io.NopCloser(buf)

		//all, err := io.ReadAll(io.NopCloser(res.Body))
		//
		//if err != nil {
		//	log.Errorf("Error: %v", err)
		//}
		////io.Teereader, io.MultiReader()
		//log.Debugf("Body\n%s\n", all)
		//
		//res.Body = io.NopCloser(bytes.NewReader(all))
	}
	return
}
