package photoprism

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

const (
	insta360PairAspectTolerance   = 0.05
	insta360PairFpsTolerance      = 0.1
	insta360PairDurationTolerance = time.Second
)

// Insta360Capture contains the original lens files and optional low-resolution proxy for one capture.
type Insta360Capture struct {
	Name  media.Insta360VideoName
	Left  *MediaFile
	Right *MediaFile
	Proxy *MediaFile
}

// FindInsta360Capture resolves the files belonging to the same directory-scoped capture as f.
func FindInsta360Capture(f *MediaFile) *Insta360Capture {
	if f == nil || !f.IsInsv() {
		return nil
	}

	name, ok := media.ParseInsta360VideoName(f.FileName())

	// The capture files are looked up under their canonical names, which must include f itself.
	if !ok || name.FileName(name.Role) != f.FileName() {
		return nil
	}

	result := &Insta360Capture{Name: name}

	for role, fileName := range map[media.Insta360VideoRole]string{
		media.Insta360VideoLeft:  name.FileName(media.Insta360VideoLeft),
		media.Insta360VideoRight: name.FileName(media.Insta360VideoRight),
		media.Insta360VideoProxy: name.FileName(media.Insta360VideoProxy),
	} {
		if !fs.FileExistsNotEmpty(fileName) {
			continue
		}

		captureFile, err := NewMediaFile(fileName)
		if err != nil {
			continue
		}

		switch role {
		case media.Insta360VideoLeft:
			result.Left = captureFile
		case media.Insta360VideoRight:
			result.Right = captureFile
		case media.Insta360VideoProxy:
			result.Proxy = captureFile
		}
	}

	return result
}

// insta360SkipConvert reports whether f is the right lens or proxy of a video capture, whose
// sidecars are created from the left lens, or an LRV proxy, which is never converted.
func insta360SkipConvert(f *MediaFile) bool {
	if f != nil && f.HasFileType(fs.VideoLrv) {
		return true
	}

	capture := FindInsta360Capture(f)
	return capture.ValidPair() && capture.Left.FileName() != f.FileName()
}

// insta360ProxyPartner returns the existing partner of a left lens video or its LRV proxy, as written by
// cameras that store both lenses in one file: the proxy for the left lens, and the left lens for the proxy.
func insta360ProxyPartner(f *MediaFile) string {
	if f == nil {
		return ""
	}

	dir, baseName := filepath.Split(f.FileName())
	var partner string

	if match := fs.Insta360ProxyPattern.FindStringSubmatch(baseName); match != nil && match[1] == "LRV" && strings.HasSuffix(baseName, fs.ExtLrv) {
		partner = fmt.Sprintf("VID_%s_%s_00_%s%s", match[2], match[3], match[5], fs.ExtInsv)
	} else if match = fs.Insta360VideoPattern.FindStringSubmatch(baseName); match != nil && match[1] == "VID" && match[4] == "00" && strings.HasSuffix(baseName, fs.ExtInsv) {
		partner = fmt.Sprintf("LRV_%s_%s_01_%s%s", match[2], match[3], match[5], fs.ExtLrv)
	} else {
		return ""
	}

	if partner = filepath.Join(dir, partner); fs.FileExistsNotEmpty(partner) {
		return partner
	}

	return ""
}

// insta360PairPreview returns the complete capture whose left lens m is the generated preview of.
func insta360PairPreview(m *MediaFile) *Insta360Capture {
	if m == nil || !m.IsPreviewImage() {
		return nil
	}

	sourceName := m.generatedSourceName()
	if fs.FileType(sourceName) != fs.VideoInsv {
		return nil
	}

	source, err := NewMediaFile(sourceName)
	if err != nil {
		return nil
	}

	if capture := FindInsta360Capture(source); capture.ValidPair() && capture.Left.FileName() == source.FileName() {
		return capture
	}

	return nil
}

// insta360RightLensSidecar reports whether m was generated from the right lens of a complete capture.
func insta360RightLensSidecar(m *MediaFile) bool {
	sourceName := m.generatedSourceName()
	if fs.FileType(sourceName) != fs.VideoInsv {
		return false
	}

	source, err := NewMediaFile(sourceName)
	if err != nil {
		return false
	}

	capture := FindInsta360Capture(source)

	return capture.ValidPair() && capture.Right.FileName() == source.FileName()
}

// insta360StalePreview reports whether the sidecar preview of a complete capture's left lens is not
// 2:1 while its right lens is among the pending files, i.e. was made before both lenses were present.
func insta360StalePreview(f *MediaFile, pending MediaFiles) bool {
	capture := FindInsta360Capture(f)
	if !capture.ValidPair() || capture.Left.FileName() != f.FileName() {
		return false
	}

	rightPending := false
	for _, file := range pending {
		if file != nil && file.FileName() == capture.Right.FileName() {
			rightPending = true
			break
		}
	}

	if !rightPending {
		return false
	}

	previewName := fs.ImageJpeg.FindFirst(f.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false)
	if previewName == "" {
		return false
	}

	preview, err := NewMediaFile(previewName)

	return err == nil && preview.InSidecar() && preview.Width() > 0 && !preview.DualFisheyeLayout()
}

// MemberPreview reports whether the specified file name is a preview of the right lens or proxy.
func (m *Insta360Capture) MemberPreview(rootRelName string) bool {
	if m == nil || rootRelName == "" {
		return false
	}

	sourceName := strings.TrimSuffix(rootRelName, filepath.Ext(rootRelName))
	for _, file := range (MediaFiles{m.Right, m.Proxy}) {
		if file != nil && file.RootRelName() == sourceName {
			return true
		}
	}

	return false
}

// ValidPair reports whether the two full-resolution lens files can safely be combined.
func (m *Insta360Capture) ValidPair() bool {
	if m == nil || m.Left == nil || m.Right == nil || !m.Left.IsInsv() || !m.Right.IsInsv() {
		return false
	}

	leftWidth, leftHeight := m.Left.Width(), m.Left.Height()
	rightWidth, rightHeight := m.Right.Width(), m.Right.Height()

	if leftWidth > 0 && leftHeight > 0 && math.Abs(float64(leftWidth)/float64(leftHeight)-1) > insta360PairAspectTolerance {
		return false
	}

	if rightWidth > 0 && rightHeight > 0 && math.Abs(float64(rightWidth)/float64(rightHeight)-1) > insta360PairAspectTolerance {
		return false
	}

	if leftWidth > 0 && rightWidth > 0 && (leftWidth != rightWidth || leftHeight != rightHeight) {
		return false
	}

	leftInfo, rightInfo := m.Left.VideoInfo(), m.Right.VideoInfo()

	if leftInfo.FPS > 0 && rightInfo.FPS > 0 && math.Abs(leftInfo.FPS-rightInfo.FPS) > insta360PairFpsTolerance {
		return false
	}

	if leftInfo.Duration > 0 && rightInfo.Duration > 0 && absDuration(leftInfo.Duration-rightInfo.Duration) > insta360PairDurationTolerance {
		return false
	}

	return true
}

// Files returns capture members in canonical lens and proxy order.
func (m *Insta360Capture) Files() MediaFiles {
	if m == nil {
		return nil
	}

	result := make(MediaFiles, 0, 3)
	for _, file := range (MediaFiles{m.Left, m.Right, m.Proxy}) {
		if file != nil {
			result = append(result, file)
		}
	}

	return result
}

// DewarpableInsv reports whether an INSV contains or belongs to a complete dual-fisheye frame.
func (m *MediaFile) DewarpableInsv() bool {
	if m == nil || !m.IsInsv() {
		return false
	}

	if capture := FindInsta360Capture(m); capture != nil && capture.ValidPair() {
		return true
	}

	return m.DualFisheyeLayout()
}

// DewarpedVideoFile returns an existing equirectangular AVC for an Insta360 video.
func DewarpedVideoFile(m *MediaFile) *MediaFile {
	if m == nil || !m.DewarpableInsv() {
		return nil
	}

	if result := m.AvcFile(); result != nil {
		return result
	}

	if capture := FindInsta360Capture(m); capture != nil && capture.Proxy != nil {
		return capture.Proxy.AvcFile()
	}

	return nil
}

// absDuration returns the absolute value of a duration.
func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}

	return value
}
