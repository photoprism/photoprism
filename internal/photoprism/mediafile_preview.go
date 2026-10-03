package photoprism

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// findPreviewImage returns the first image of the given types named after fileName, in its own folder,
// the sidecar folder, or the hidden folder, whose content matches its type.
func findPreviewImage(fileName, sidecarPath, originalsPath string, stripSequence bool, types ...fs.Type) *MediaFile {
	if fileName == "" {
		return nil
	}

	dirs := []string{sidecarPath, fs.PPHiddenPathname}
	checked := make(map[string]bool)

	for _, fileType := range types {
		for _, name := range fileType.FindAll(fileName, dirs, originalsPath, stripSequence) {
			if checked[name] {
				continue
			}

			checked[name] = true

			if !previewContentMatches(name, fileType) {
				log.Debugf("media: %s is not used as preview, as its content does not match its type", clean.Log(filepath.Base(name)))
			} else if preview, err := NewMediaFile(name); err == nil && preview.Ok() && preview.IsPreviewImage() {
				return preview
			}
		}
	}

	return nil
}

// previewContentMatches reports whether the file header matches the JPEG or PNG type of a preview
// image. It reads the header only, unlike MediaFile.MimeType, which may read metadata for videos.
func previewContentMatches(fileName string, fileType fs.Type) bool {
	switch fileType {
	case fs.ImageJpeg:
		return fs.SameType(fs.MimeType(fileName), header.ContentTypeJpeg)
	case fs.ImagePng:
		return header.IsPngType(fs.MimeType(fileName))
	default:
		return false
	}
}
