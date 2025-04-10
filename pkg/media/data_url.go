package media

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// DataUrl returns a data URL representing the binary buffer data.
func DataUrl(buf *bytes.Buffer) string {
	encoded := EncodeBase64(buf.Bytes())

	if encoded == "" {
		return ""
	}

	var mimeType string

	mime, err := mimetype.DetectReader(buf)

	if err != nil {
		mimeType = "application/octet-stream"
	} else {
		mimeType = mime.String()
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// ReadUrl reads binary data from a regular file path,
// fetches its data from a remote http or https URL,
// or decodes a base64 data URL as created by DataUrl.
func ReadUrl(file string) (data []byte, err error) {
	u, err := url.Parse(file)
	if err != nil {
		log.Fatal(err)
	}

	// Also supports http, https, and data URLs instead of a file name for remote processing.
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, httpErr := http.Get(file)

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
			return DecodeBase64(binaryData)
		}
	} else if data, err = os.ReadFile(file); err != nil {
		return nil, err
	}

	return data, err
}
