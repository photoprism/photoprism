package photoprism

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/list"
)

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles(stripSequence bool) (result RelatedFiles, err error) {
	// Related file path prefix without ignored file name extensions and suffixes.
	filePathPrefix := m.AbsPrefix(stripSequence)

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

	// Find additional sidecar files with naming schemes not matching the glob pattern,
	// see https://github.com/photoprism/photoprism/issues/2983 for further information.
	if files, _ := m.RelatedSidecarFiles(stripSequence); len(files) > 0 {
		matches = list.Join(matches, files)
	}

	isHeic := false

	// Process files that matched the pattern.
	for _, fileName := range matches {
		f, fileErr := NewMediaFile(fileName)

		if fileErr != nil || f.Empty() || f.IsArchive() {
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
		switch {
		case result.Main == nil && f.IsPreviewImage():
			result.Main = f
		case f.IsRaw():
			result.Main = f
		case f.IsVector():
			result.Main = f
		case f.IsDocument():
			result.Main = f
		case f.IsHeic():
			isHeic = true
			result.Main = f
		case f.IsHeif():
			result.Main = f
		case f.IsImage() && !f.IsPreviewImage() && !f.IsThumb():
			result.Main = f
		case f.IsVideo() && !isHeic:
			result.Main = f
		case result.Main != nil && f.IsPreviewImage() && result.Main.IsPreviewImage() && len(result.Main.FileName()) > len(f.FileName()):
			result.Main = f
		}

		result.Files = append(result.Files, f)
	}

	if len(result.Files) == 0 || result.Main == nil {
		mediaType := m.BaseType()

		if mediaType == "" {
			mediaType = "unknown type"
		}

		return result, fmt.Errorf("%s is unsupported (%s)", clean.Log(m.BaseName()), mediaType)
	}

	// Add hidden preview image if needed.
	if !result.HasPreview() {
		if jpegName := fs.ImageJpeg.FindFirst(result.Main.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), stripSequence); jpegName != "" {
			if resultFile, _ := NewMediaFile(jpegName); resultFile.Ok() {
				result.Files = append(result.Files, resultFile)
			}
		} else if pngName := fs.ImagePng.FindFirst(result.Main.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), stripSequence); pngName != "" {
			if resultFile, _ := NewMediaFile(pngName); resultFile.Ok() {
				result.Files = append(result.Files, resultFile)
			}
		}
	}

	sort.Sort(result.Files)

	return result, nil
}

// RelatedSidecarFiles finds additional sidecar files with naming schemes not matching the default glob pattern
// for related files. see https://github.com/photoprism/photoprism/issues/2983 for further information.
func (m *MediaFile) RelatedSidecarFiles(stripSequence bool) (files []string, err error) {
	baseName := filepath.Base(m.fileName)
	files = make([]string, 0, 2)

	// Find edited file versions with a naming scheme as used by Apple, for example "IMG_E12345.JPG".
	if strings.ToUpper(baseName[:4]) == "IMG_" && strings.ToUpper(baseName[:5]) != "IMG_E" {
		if fileName := filepath.Join(filepath.Dir(m.fileName), baseName[:4]+"E"+baseName[4:]); fs.FileExists(fileName) {
			files = append(files, fileName)
		}
	}

	// Related file path prefix without ignored file name extensions and suffixes.
	filePathPrefix := m.AbsPrefix(stripSequence)

	// Find additional sidecar files that match the default glob pattern for related files.
	globPattern := regexp.QuoteMeta(filePathPrefix) + "_????\\.*"
	matches, err := filepath.Glob(globPattern)

	if err != nil {
		return files, err
	}

	// Add glob file matches to results.
	files = append(files, matches...)

	return files, nil
}
