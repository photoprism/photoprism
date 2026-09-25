package media

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"
)

// FromName returns the content type matching the file extension.
func FromName(fileName string) Type {
	if fileName == "" {
		return Unknown
	}

	fileType := fs.FileType(fileName)

	// Proxies of other cameras share the extension of Insta360 proxies, but are not supported.
	if fileType == fs.VideoLrv && !fs.Insta360ProxyPattern.MatchString(filepath.Base(fileName)) {
		return Sidecar
	}

	// Find media type based on the file type.
	if result, found := Formats[fileType]; found {
		return result
	}

	// Default to sidecar.
	return Sidecar
}

// MainFile checks if the filename belongs to a main content type. Proxy videos are not, since
// they are only indexed with the video they belong to.
func MainFile(fileName string) bool {
	return fs.FileType(fileName) != fs.VideoLrv && FromName(fileName).IsMain()
}
