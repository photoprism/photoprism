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

// RelatedFilePathPrefix returns the absolute file path and name prefix without file extensions
// and suffixes to be ignored.
func (m *MediaFile) RelatedFilePathPrefix(stripSequence bool) (s string) {
	return fs.RelatedFilePathPrefix(m.FileName(), stripSequence)
}

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles(stripSequence bool) (result RelatedFiles, err error) {
	// Related file path prefix without ignored file name extensions and suffixes.
	filePathPrefix := m.RelatedFilePathPrefix(stripSequence)

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
	if len(sidecarPrefix) > 1 && sidecarPrefix != originalsPrefix && strings.HasPrefix(filePathPrefix, sidecarPrefix) {
		filePathPrefix = strings.Replace(filePathPrefix, sidecarPrefix, originalsPrefix, 1)
		log.Debugf("media: replaced sidecar with originals path in related file matching pattern")
	}

	// globPattern specifies the escaped naming pattern to find related files.
	var globPattern string

	// Strip common name sequences like "copy 2" or "(3)"?
	if stripSequence {
		globPattern = regexp.QuoteMeta(filePathPrefix) + "*"
	} else {
		globPattern = regexp.QuoteMeta(filePathPrefix+".") + "*"
	}

	// Find files that match the pattern.
	matches, err := filepath.Glob(globPattern)

	if err != nil {
		return result, err
	}

	// Additionally include edited version in the file matches, if exists.
	if name := m.EditedName(); name != "" {
		matches = append(matches, name)
	}

	isHEIC := false

	// Process files that matched the pattern.
	for _, fileName := range matches {
		f, fileErr := NewMediaFile(fileName)

		if fileErr != nil || f.Empty() {
			continue
		}

		// Skip file if its format must be ignored based on the configuration.
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
