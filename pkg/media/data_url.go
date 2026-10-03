package media

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/safe"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

const imageAcceptHeader = "image/jpeg, image/png, image/webp, image/avif, image/heic, image/heif, */*;q=0.1"

// ErrImageTooLarge is returned when image data exceeds MaxImageBytes.
var ErrImageTooLarge = errors.New("image data exceeds the supported size")

var (
	// MaxImageBytes limits image data read from a URL. It is sized for a full-resolution frame
	// in the formats the Accept header requests, not for a generic file download.
	MaxImageBytes int64 = 32 << 20

	// ImageReadTimeout limits how long reading image data from a remote URL may take.
	ImageReadTimeout = 30 * time.Second
)

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

// imageDownloadOptions returns the transfer options used for reading image data from a URL.
// They are stated here rather than inherited, so an image reference cannot deliver as much as
// a generic file download may.
func imageDownloadOptions() *safe.Options {
	return &safe.Options{
		AllowPrivate: false,
		Accept:       imageAcceptHeader,
		MaxSizeBytes: MaxImageBytes,
		Timeout:      ImageReadTimeout,
	}
}

// ReadUrlImage reads binary image data with strict remote URL safety defaults and an image
// budget, so that an image reference cannot deliver as much as a generic download may.
func ReadUrlImage(fileUrl string, schemes []string) (data []byte, err error) {
	data, err = ReadUrlWithOptions(fileUrl, schemes, imageDownloadOptions())

	if err != nil {
		return data, err
	}

	// Also applies to the schemes that carry their data inline, which the download options
	// above never see.
	if MaxImageBytes > 0 && int64(len(data)) > MaxImageBytes {
		return nil, fmt.Errorf("%w (%d bytes)", ErrImageTooLarge, len(data))
	}

	return data, nil
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
		} else if opt != nil && opt.MaxSizeBytes > 0 && int64(len(binaryData))/4*3 > opt.MaxSizeBytes {
			// Checked on the encoded length, so an oversized payload is refused before it is
			// decoded rather than after.
			return data, fmt.Errorf("%w (%d bytes)", ErrImageTooLarge, len(binaryData)/4*3)
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
