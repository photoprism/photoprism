package photoprism

import (
	"path"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// xmpMainExts lists the canonical lowercase main-media extensions used for XMP main-file lookup.
var xmpMainExts = media.MainExtensions()

// mainForSidecar resolves the main media file for an XMP sidecar from the originals Files cache.
func (ind *Index) mainForSidecar(relName string) string {
	// Full-name convention: "tok/photo.jpg.xmp" -> "tok/photo.jpg".
	if main := fs.StripExt(relName); media.MainFile(main) {
		if ind.files.Exists(main, entity.RootOriginals) && ind.sidecarMainEnabled(main) {
			return main
		}

		return ""
	}

	// Base-prefix convention: "tok/photo.xmp" -> "tok/photo.<mainext>". The main file's extension
	// is unknown, so each main extension is probed against the Files cache in both lowercase and
	// uppercase form (the cache preserves each file's on-disk case). The base name must match case.
	dir := path.Dir(relName)
	base := fs.BasePrefix(relName, false)

	for _, ext := range xmpMainExts {
		for _, name := range []string{base + ext, base + strings.ToUpper(ext)} {
			candidate := path.Join(dir, name)

			if ind.files.Exists(candidate, entity.RootOriginals) && ind.sidecarMainEnabled(candidate) {
				return candidate
			}
		}
	}

	return ""
}

// sidecarMainEnabled reports whether a cached main media file is enabled by the current config.
func (ind *Index) sidecarMainEnabled(relName string) bool {
	fileType := fs.FileType(relName)

	if !media.Formats[fileType].IsMain() {
		return false
	}

	switch {
	case (fileType == fs.ImageRaw || fileType == fs.ImageDng) && ind.conf.DisableRaw():
		return false
	case fileType == fs.ImageJpegXL && ind.conf.DisableJpegXL():
		return false
	case media.Formats[fileType] == media.Vector && ind.conf.DisableVectors():
		return false
	default:
		return true
	}
}
