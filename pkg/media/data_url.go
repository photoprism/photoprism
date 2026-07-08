package media

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/safe"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

const imageAcceptHeader = "image/jpeg, image/png, image/webp, image/avif, image/heic, image/heif, */*;q=0.1"

// DataUrl generates a data URL of the binary data from the specified io.Reader.
func DataUrl(r io.Reader) string {
	// Read binary data.
	data, err := io.ReadAll(r)

	if err != nil || len(data) == 0 {
		return ""
	}

	// Return as string if it already appears to be a data URL.
	if string(data[0:4]) == "data:" {
		return string(data)
	}

	// Detect mime type.
	var mime *mimetype.MIME
	var mimeType string

	if mime = mimetype.Detect(data); mime == nil {
		mimeType = header.ContentTypeBinary
	} else {
		mimeType = mime.String()
	}

	// Generate data URL.
	return fmt.Sprintf("data:%s;base64,%s", mimeType, EncodeBase64String(data))
}

// DataBase64 generates a base64 encoded string of the binary data from the specified io.Reader.
func DataBase64(r io.Reader) string {
	// Read binary data.
	data, err := io.ReadAll(r)

	if err != nil || len(data) == 0 {
		return ""
	}

	return EncodeBase64String(data)
}

// ReadUrl reads binary data from a regular file path,
// fetches its data from a remote http or https URL,
// or decodes a base64 data URL as created by DataUrl.
func ReadUrl(fileUrl string, schemes []string) (data []byte, err error) {
	return ReadUrlWithOptions(fileUrl, schemes, nil)
}

// ReadUrlImage reads binary image data with strict remote URL safety defaults.
func ReadUrlImage(fileUrl string, schemes []string) (data []byte, err error) {
	return ReadUrlWithOptions(fileUrl, schemes, &safe.Options{
		AllowPrivate: false,
		Accept:       imageAcceptHeader,
	})
}

// ReadUrlWithOptions reads binary data while applying optional safe HTTP options for remote URLs.
func ReadUrlWithOptions(fileUrl string, schemes []string, opt *safe.Options) (data []byte, err error) {
	if fileUrl == "" {
		return data, errors.New("missing url")
	}

	// Parse file URL.
	var u *url.URL

	if u, err = url.Parse(fileUrl); err != nil {
		return data, fmt.Errorf("invalid url (%s)", err)
	}

	// Reject it if it is not absolute, i.e. it does not contain a scheme.
	if !u.IsAbs() {
		return data, fmt.Errorf("url %s requires a scheme", clean.Log(fileUrl))
	} else if !slices.Contains(schemes, u.Scheme) {
		return data, fmt.Errorf("invalid url scheme %s", clean.Log(u.Scheme))
	}

	// Fetch the file data from the specified URL, depending on its scheme.
	switch u.Scheme {
	case scheme.Https, scheme.Http:
		if data, err = readRemoteUrl(fileUrl, opt); err != nil {
			return data, fmt.Errorf("invalid %s url (%w)", u.Scheme, err)
		}
	case scheme.Unix, scheme.HttpUnix:
		return data, fmt.Errorf("unsupported url scheme %s", clean.Log(u.Scheme))
	case scheme.Data:
		if _, binaryData, found := strings.Cut(u.Opaque, ";base64,"); !found || len(binaryData) == 0 {
			return data, fmt.Errorf("invalid %s url", u.Scheme)
		} else {
			return DecodeBase64String(binaryData)
		}
	case scheme.File:
		path := u.Path
		if path == "" {
			path = u.Opaque
		}
		if path == "" {
			return data, fmt.Errorf("invalid %s url (empty path)", u.Scheme)
		}
		if data, err = os.ReadFile(path); err != nil { //nolint:gosec // file path validated earlier
			return data, fmt.Errorf("invalid %s url (%s)", u.Scheme, err)
		}
	default:
		return data, fmt.Errorf("unsupported url scheme %s", clean.Log(u.Scheme))
	}

	return data, err
}

// readRemoteUrl downloads a remote URL with safe defaults and returns the resulting bytes.
func readRemoteUrl(rawURL string, opt *safe.Options) (data []byte, err error) {
	tmpFile, err := os.CreateTemp("", "photoprism-read-url-*")
	if err != nil {
		return data, err
	}

	tmpName := tmpFile.Name()

	if closeErr := tmpFile.Close(); closeErr != nil {
		return data, closeErr
	}

	defer func() {
		_ = os.Remove(tmpName)
	}()

	options := &safe.Options{
		AllowPrivate: true,
		Accept:       "*/*",
	}

	if opt != nil {
		options = &safe.Options{
			Timeout:      opt.Timeout,
			MaxSizeBytes: opt.MaxSizeBytes,
			AllowPrivate: opt.AllowPrivate,
			Accept:       opt.Accept,
		}
	}

	if err = safe.Download(tmpName, rawURL, options); err != nil {
		return data, err
	}

	data, err = os.ReadFile(tmpName) //nolint:gosec // tmpName is created by os.CreateTemp

	return data, err
}
