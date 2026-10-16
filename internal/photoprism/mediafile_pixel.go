package photoprism

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GooglePixelCapture contains the files belonging to a Google Pixel Camera multi-file capture.
type GooglePixelCapture struct {
	MainPhoto *MediaFile // Primary photo (processed JPEG or composite)
	MainVideo *MediaFile // Primary boosted video
	Originals MediaFiles // Stacked originals (companion RAW files, unenhanced JPEGs, draft videos)
}

// FindGooglePixelCapture resolves the files belonging to the same Google Pixel Camera capture as f.
func FindGooglePixelCapture(f *MediaFile) *GooglePixelCapture {
	if f == nil || !f.IsMedia() || f.IsSidecar() || f.InSidecar() {
		return nil
	}

	basePrefix := f.BasePrefix(false)
	if !fs.GooglePixelPattern.MatchString(basePrefix) {
		return nil
	}

	stackPrefix := f.StackPrefix(false)
	if stackPrefix == "" {
		return nil
	}

	dir := filepath.Dir(f.FileName())
	pattern := filepath.Join(dir, stackPrefix+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil
	}

	capture := &GooglePixelCapture{
		Originals: make(MediaFiles, 0, len(matches)),
	}

	var draftVideo *MediaFile

	for _, fileName := range matches {
		mf, mfErr := NewMediaFile(fileName)
		if mfErr != nil || mf.Empty() || !mf.IsMedia() || mf.IsSidecar() || mf.InSidecar() {
			continue
		}

		if !fs.GooglePixelPattern.MatchString(mf.BasePrefix(false)) {
			continue
		}

		upper := strings.ToUpper(mf.BasePrefix(false))

		if mf.IsVideo() {
			if strings.Contains(upper, ".MAIN") {
				// If multiple MAIN videos exist (e.g. 02.MAIN and 03.MAIN), prefer the higher sequence (03 > 02).
				if capture.MainVideo == nil || mf.BasePrefix(false) > capture.MainVideo.BasePrefix(false) {
					if capture.MainVideo != nil {
						capture.Originals = append(capture.Originals, capture.MainVideo)
					}
					capture.MainVideo = mf
				} else {
					capture.Originals = append(capture.Originals, mf)
				}
			} else if strings.Contains(upper, ".COVER") || strings.Contains(upper, "-01") {
				if draftVideo != nil {
					capture.Originals = append(capture.Originals, draftVideo)
				}
				draftVideo = mf
			} else {
				capture.Originals = append(capture.Originals, mf)
			}
		} else {
			// Photo capture: classify into MainPhoto (primary) vs Originals (stacked related files).
			// The primary photo is the processed JPEG (-01, .COVER, or .PORTRAIT).
			// Companion RAW files, .ORIGINAL unenhanced shots, and secondary frames (-02, -03, etc.) are stacked originals.
			isMain := !mf.IsRaw() && !strings.Contains(upper, ".ORIGINAL") && (strings.Contains(upper, "-01") || strings.Contains(upper, ".COVER") || strings.HasSuffix(upper, ".PORTRAIT"))

			if isMain {
				if capture.MainPhoto != nil {
					capture.Originals = append(capture.Originals, capture.MainPhoto)
				}
				capture.MainPhoto = mf
			} else {
				capture.Originals = append(capture.Originals, mf)
			}
		}
	}

	// For video captures, if a MainVideo was found, the draft preview video is stacked as an original.
	// If no MainVideo exists yet (e.g. still uploading/processing in Google cloud), the draft preview serves as MainVideo.
	if draftVideo != nil {
		if capture.MainVideo != nil {
			capture.Originals = append(capture.Originals, draftVideo)
		} else {
			capture.MainVideo = draftVideo
		}
	}

	return capture
}

// Primary returns the main viewable file: for videos it returns the boosted MainVideo (falling back to draft preview),
// and for photos it returns MainPhoto.
func (c *GooglePixelCapture) Primary() *MediaFile {
	if c == nil {
		return nil
	}
	if c.MainVideo != nil {
		return c.MainVideo
	}
	if c.MainPhoto != nil {
		return c.MainPhoto
	}
	if len(c.Originals) > 0 {
		return c.Originals[0]
	}
	return nil
}

// ValidPair reports whether the capture contains a primary file and at least one stacked original file.
func (c *GooglePixelCapture) ValidPair() bool {
	return c != nil && c.Primary() != nil && len(c.Files()) > 1
}

// Files returns all member files belonging to the capture, with Primary first.
func (c *GooglePixelCapture) Files() MediaFiles {
	if c == nil {
		return nil
	}
	primary := c.Primary()
	res := make(MediaFiles, 0, 2+len(c.Originals))
	seen := make(map[string]bool, 2+len(c.Originals))

	if primary != nil {
		res = append(res, primary)
		seen[primary.FileName()] = true
	}

	for _, f := range []*MediaFile{c.MainVideo, c.MainPhoto} {
		if f != nil && !seen[f.FileName()] {
			res = append(res, f)
			seen[f.FileName()] = true
		}
	}

	for _, f := range c.Originals {
		if f != nil && !seen[f.FileName()] {
			res = append(res, f)
			seen[f.FileName()] = true
		}
	}

	return res
}

// IsPrimary reports whether f is the primary file of the capture.
func (c *GooglePixelCapture) IsPrimary(f *MediaFile) bool {
	if c == nil || f == nil {
		return false
	}
	primary := c.Primary()
	return primary != nil && primary.FileName() == f.FileName()
}

// IsMember reports whether a relative or absolute file name belongs to this capture.
func (c *GooglePixelCapture) IsMember(fileName string) bool {
	if c == nil || fileName == "" {
		return false
	}
	base := filepath.Base(fileName)
	for _, f := range c.Files() {
		if f == nil {
			continue
		}
		if f.FileName() == fileName || f.RootRelName() == fileName || f.BaseName() == base {
			return true
		}
	}
	return false
}

// MemberFile reports whether file is a related original or RAW sidecar of this capture.
func (c *GooglePixelCapture) MemberFile(file entity.File) bool {
	if c == nil {
		return false
	}
	return file.FileSidecar || file.FileRoot == entity.RootSidecar || c.IsMember(file.FileName)
}

// MemberPreview reports whether fileName is a generated preview of a secondary member of this capture (e.g. draft video preview).
func (c *GooglePixelCapture) MemberPreview(fileName string) bool {
	if c == nil || fileName == "" {
		return false
	}
	base := filepath.Base(fileName)
	for _, f := range c.Files() {
		if f == nil || c.IsPrimary(f) {
			continue
		}
		if f.BaseName()+".jpg" == base || strings.HasPrefix(base, f.BaseName()) {
			return true
		}
	}
	return false
}

// googlePixelSkipConvert reports whether f is a companion RAW file belonging to a valid Google Pixel Camera
// capture whose primary image already exists, avoiding duplicate sidecars.
func googlePixelSkipConvert(f *MediaFile) bool {
	if f == nil || !f.IsRaw() {
		return false
	}
	capture := FindGooglePixelCapture(f)
	if !capture.ValidPair() {
		return false
	}
	primary := capture.Primary()
	return primary != nil && (primary.IsJpeg() || primary.IsHeif() || primary.HasPreviewImage())
}

// googlePixelImportOrder returns the related files in the order they are imported, with the primary
// image or video first, so it keeps the unsuffixed name that the other files of its set are named after.
func googlePixelImportOrder(related RelatedFiles) MediaFiles {
	if capture := FindGooglePixelCapture(related.Main); !capture.ValidPair() || !capture.IsPrimary(related.Main) {
		return related.Files
	}

	result := make(MediaFiles, 0, len(related.Files))
	result = append(result, related.Main)

	for _, f := range related.Files {
		if f != nil && f.FileName() != related.Main.FileName() {
			result = append(result, f)
		}
	}

	return result
}

// googlePixelPrimaryCapture returns the capture if m is the canonical primary file of a valid multi-file capture.
func googlePixelPrimaryCapture(m *MediaFile) *GooglePixelCapture {
	if capture := FindGooglePixelCapture(m); capture != nil && capture.ValidPair() && capture.IsPrimary(m) {
		return capture
	}
	return nil
}

// googlePixelPairPreview returns the complete capture whose boosted video m is the generated preview of.
func googlePixelPairPreview(m *MediaFile) *GooglePixelCapture {
	if m == nil || !m.IsPreviewImage() {
		return nil
	}

	sourceName := m.generatedSourceName()
	if sourceName == "" || !fs.GooglePixelPattern.MatchString(filepath.Base(sourceName)) {
		return nil
	}

	source, err := NewMediaFile(sourceName)
	if err != nil {
		return nil
	}

	if capture := FindGooglePixelCapture(source); capture != nil && capture.ValidPair() && capture.IsPrimary(source) {
		return capture
	}

	return nil
}
