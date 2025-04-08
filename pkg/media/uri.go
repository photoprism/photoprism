package media

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ReadUri reads binary data from regular files, fetches data from remote http and
// https URLs, or decodes base64 data URLs - depending on the type of URI you pass.
func ReadUri(uri string) (data []byte, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal(err)
	}

	// Also supports http, https, and data URLs instead of a file name for remote processing.
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, httpErr := http.Get(uri)

		if httpErr != nil {
			return nil, httpErr
		}

		defer resp.Body.Close()

		if data, err = io.ReadAll(resp.Body); err != nil {
			return nil, err
		}
	} else if u.Scheme == "data" {
		if _, binaryData, found := strings.Cut(u.Opaque, ";base64,"); !found || len(binaryData) == 0 {
			return nil, fmt.Errorf("invalid data URL")
		} else {
			return base64.StdEncoding.DecodeString(binaryData)
		}
	} else if data, err = os.ReadFile(uri); err != nil {
		return nil, err
	}

	return data, err
}
