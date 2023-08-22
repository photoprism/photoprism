package photoprism

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles(stripSequence bool) (result RelatedFiles, err error) {
	// File path and name without any extensions.
	prefix := m.AbsPrefix(stripSequence)

	// Storage folder path prefixes.
	sidecarPrefix := Config().SidecarPath() + "/"
	originalsPrefix := Config().OriginalsPath() + "/"

	// Ignore RAW images?
	skipRaw := Config().DisableRaw()

	// Ignore JPEG XL files?
	skipJpegXL := Config().DisableJpegXL()

	// Ignore vector graphics?
	skipVectors := Config().DisableVectors()

	// Replace sidecar with originals path in search prefix.
	if len(sidecarPrefix) > 1 && sidecarPrefix != originalsPrefix && strings.HasPrefix(prefix, sidecarPrefix) {
		prefix = strings.Replace(prefix, sidecarPrefix, originalsPrefix, 1)
		log.Debugf("media: replaced sidecar with originals path in related file matching pattern")
	}

	// Quote path for glob.
	if stripSequence {
		// Strip common name sequences like "copy 2" and escape meta characters.
		prefix = regexp.QuoteMeta(prefix)
	} else {
		// Use strict file name matching and escape meta characters.
		prefix = regexp.QuoteMeta(prefix + ".")
	}

	// Find related files.
	matches, err := filepath.Glob(prefix + "*")

	if err != nil {
		return result, err
	}

	// Search for related edited image file name (as used by Apple) and add it to the list of files, if found.
	if name := m.EditedName(); name != "" {
		matches = append(matches, name)
	}

	// Extract an embedded video file and add it to the list of files, if successful.
	if videoName, videoErr := m.ExtractEmbeddedVideo(); videoErr != nil {
		log.Warnf("media: %s om %s (extract embedded video)", clean.Error(videoErr), clean.Log(m.RootRelName()))
	} else if videoName != "" {
		matches = append(matches, videoName)
	}

	isHEIC := false

	for _, fileName := range matches {
		f, fileErr := NewMediaFile(fileName)

		if fileErr != nil || f.Empty() {
			continue
		}

		// Ignore file format?
		switch {
		case skipRaw && f.IsRaw():
			log.Debugf("media: skipped related raw image %s", clean.Log(f.RootRelName()))
			continue
		case skipJpegXL && f.IsJpegXL():
			log.Debugf("media: skipped related JPEG XL file %s", clean.Log(f.RootRelName()))
			continue
		case skipVectors && f.IsVector():
			log.Debugf("media: skipped related vector graphic %s", clean.Log(f.RootRelName()))
			continue
		}

		// Set main file.
		if result.Main == nil && f.IsPreviewImage() {
			result.Main = f
		} else if f.IsRaw() {
			result.Main = f
		} else if f.IsVector() {
			result.Main = f
		} else if f.IsHEIC() {
			isHEIC = true
			result.Main = f
		} else if f.IsHEIF() {
			result.Main = f
		} else if f.IsImage() && !f.IsPreviewImage() {
			result.Main = f
		} else if f.IsVideo() && !isHEIC {
			result.Main = f
		} else if result.Main != nil && f.IsPreviewImage() {
			if result.Main.IsPreviewImage() && len(result.Main.FileName()) > len(f.FileName()) {
				result.Main = f
			}
		}

		result.Files = append(result.Files, f)
	}

	if len(result.Files) == 0 || result.Main == nil {
		t := m.MimeType()

		if t == "" {
			t = "unknown type"
		}

		return result, fmt.Errorf("%s is unsupported (%s)", clean.Log(m.BaseName()), t)
	}

	// Add hidden preview image if needed.
	if !result.HasPreview() {
		if jpegName := fs.ImageJPEG.FindFirst(result.Main.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), stripSequence); jpegName != "" {
			if resultFile, _ := NewMediaFile(jpegName); resultFile.Ok() {
				result.Files = append(result.Files, resultFile)
			}
		} else if pngName := fs.ImagePNG.FindFirst(result.Main.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), stripSequence); pngName != "" {
			if resultFile, _ := NewMediaFile(pngName); resultFile.Ok() {
				result.Files = append(result.Files, resultFile)
			}
		}
	}

	sort.Sort(result.Files)

	return result, nil
}
