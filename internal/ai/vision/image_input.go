package vision

import (
	"bytes"
	"errors"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

// ErrNotAnImage is returned when a resolved image reference does not contain image data.
var ErrNotAnImage = errors.New("reference does not contain an image")

// remoteImageData fetches and validates a remote image reference and returns its bytes.
//
// The reference MUST be an https or data URL: media.ReadUrlImage rejects local file paths,
// file: URLs, plain http, and private/loopback targets, so a caller-supplied reference can
// never read a local file. Non-image payloads are rejected as well.
func remoteImageData(ref string) ([]byte, error) {
	data, err := media.ReadUrlImage(ref, scheme.HttpsData)

	if err != nil {
		return nil, err
	}

	if mime := mimetype.Detect(data); mime == nil || !strings.HasPrefix(mime.String(), "image/") {
		return nil, ErrNotAnImage
	}

	return data, nil
}

// remoteImageDataUrl resolves a remote image reference to an inline data URL.
func remoteImageDataUrl(ref string) (string, error) {
	data, err := remoteImageData(ref)

	if err != nil {
		return "", err
	}

	return media.DataUrl(bytes.NewReader(data)), nil
}

// remoteImageBase64 resolves a remote image reference to raw base64-encoded image data.
func remoteImageBase64(ref string) (string, error) {
	data, err := remoteImageData(ref)

	if err != nil {
		return "", err
	}

	return media.DataBase64(bytes.NewReader(data)), nil
}
