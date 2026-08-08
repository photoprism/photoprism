package photoprism

import (
	"math"
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
	if !ok {
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
