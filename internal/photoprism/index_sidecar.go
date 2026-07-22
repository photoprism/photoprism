package photoprism

import (
	"path"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// xmpPrimaryExts lists sorted main-media extensions for XMP primary lookup.
var xmpPrimaryExts = mainFileExts()

// mainFileExts returns sorted lowercase and uppercase main-media extensions.
func mainFileExts() []string {
	known := make(map[string]struct{}, len(fs.Extensions)*2)

	for ext, fileType := range fs.Extensions {
		if media.Formats[fileType].IsMain() {
			known[strings.ToLower(ext)] = struct{}{}
			known[strings.ToUpper(ext)] = struct{}{}
		}
	}

	exts := make([]string, 0, len(known))

	for ext := range known {
		exts = append(exts, ext)
	}

	sort.Strings(exts)

	return exts
}

// primaryForSidecar resolves an XMP primary from the originals Files cache.
func (ind *Index) primaryForSidecar(relName string) string {
	// Full-name convention: "tok/photo.jpg.xmp" -> "tok/photo.jpg".
	if primary := fs.StripExt(relName); media.MainFile(primary) {
		if ind.files.Exists(primary, entity.RootOriginals) && ind.sidecarPrimaryEnabled(primary) {
			return primary
		}

		return ""
	}

	// Base-prefix convention: "tok/photo.xmp" -> "tok/photo.<mainext>".
	dir := path.Dir(relName)
	base := fs.BasePrefix(relName, false)
	bases := []string{base, strings.ToLower(base), strings.ToUpper(base)}
	seen := make(map[string]struct{}, len(bases))

	for _, ext := range xmpPrimaryExts {
		for _, candidateBase := range bases {
			if _, ok := seen[candidateBase+ext]; ok {
				continue
			}

			seen[candidateBase+ext] = struct{}{}
			candidate := path.Join(dir, candidateBase+ext)

			if ind.files.Exists(candidate, entity.RootOriginals) && ind.sidecarPrimaryEnabled(candidate) {
				return candidate
			}
		}
	}

	return ""
}

// sidecarPrimaryEnabled reports whether a cached primary is enabled by the current config.
func (ind *Index) sidecarPrimaryEnabled(relName string) bool {
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
