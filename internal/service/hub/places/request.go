package places

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// GetRequest fetches the cell ID data from the service URL.
func GetRequest(reqUrl string, locale string) (r *http.Response, err error) {
	var req *http.Request

	// Log request URL.
	log.Tracef("places: sending request to %s", reqUrl)

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
		req.Header.Set("X-Signature", fmt.Sprintf("%x", sha1.Sum([]byte(Key+reqUrl+Secret))))
	}

	// Create new http.Client.
	//
	// NOTE: Timeout specifies a time limit for requests made by
	// this Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after GetRequest, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	client := &http.Client{Timeout: 60 * time.Second}

	// Perform request.
	for i := 0; i < Retries; i++ {
		r, err = client.Do(req)

		// Ok?
		if err == nil {
			return r, nil
		}

		// Wait before trying again?
		if RetryDelay.Nanoseconds() > 0 {
			time.Sleep(RetryDelay)
		}
	}

	return r, err
}
