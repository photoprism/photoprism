package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// writeInsta360CaptureFile copies a square image fixture to an INSV capture filename for geometry tests.
func writeInsta360CaptureFile(t *testing.T, dir, name, fixture string) string {
	t.Helper()
	require.NoError(t, fs.MkdirAll(dir))

	// #nosec G304 -- the fixture path is controlled by the test.
	payload, err := os.ReadFile(fixture)
	require.NoError(t, err)

	fileName := filepath.Join(dir, name)
	// #nosec G703 -- the destination directory and filename are controlled by the test.
	require.NoError(t, os.WriteFile(fileName, payload, fs.ModeFile))

	return fileName
}

// TestFindInsta360Capture verifies exact capture grouping and incomplete-pair fallback.
func TestFindInsta360Capture(t *testing.T) {
	t.Run("CompleteWithProxy", func(t *testing.T) {
		dir := t.TempDir()
		leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, dir, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg")

		left, err := NewMediaFile(leftName)
		require.NoError(t, err)
		capture := FindInsta360Capture(left)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Len(t, capture.Files(), 3)
		assert.Equal(t, media.Insta360VideoLeft, capture.Name.Role)
	})
	t.Run("Incomplete", func(t *testing.T) {
		dir := t.TempDir()
		left, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg"))
		require.NoError(t, err)
		capture := FindInsta360Capture(left)
		require.NotNil(t, capture)
		assert.False(t, capture.ValidPair())
		assert.Len(t, capture.Files(), 1)
	})
	t.Run("Unrelated", func(t *testing.T) {
		file, err := NewMediaFile("testdata/insta360.insv")
		require.NoError(t, err)
		assert.Nil(t, FindInsta360Capture(file))
	})
	t.Run("Photo", func(t *testing.T) {
		// Photos are never split by lens, so photo files with lens codes are not captures.
		dir := t.TempDir()
		left, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg"))
		require.NoError(t, err)
		right, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg"))
		require.NoError(t, err)

		assert.Nil(t, FindInsta360Capture(left))
		assert.Nil(t, FindInsta360Capture(right))
	})
	t.Run("PhotoSingleFileCaptureName", func(t *testing.T) {
		dir := t.TempDir()
		fileName := writeInsta360CaptureFile(t, dir, "IMG_20231015_101112_00_123.insp", "testdata/insta360.insp")
		file, err := NewMediaFile(fileName)
		require.NoError(t, err)

		assert.Nil(t, FindInsta360Capture(file))

		related, err := file.RelatedFiles(false)
		require.NoError(t, err)
		assert.Len(t, related.Files, 1)
		assert.Equal(t, fileName, related.Main.FileName())
	})
	t.Run("CaseVariant", func(t *testing.T) {
		dir := t.TempDir()
		writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
		variant, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "vid_20220625_140410_00_008.insv", "testdata/flash.jpg"))
		require.NoError(t, err)
		upperExt, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.INSV", "testdata/flash.jpg"))
		require.NoError(t, err)

		assert.Nil(t, FindInsta360Capture(variant))
		assert.Nil(t, FindInsta360Capture(upperExt))
	})
	t.Run("PhotoSingleFile", func(t *testing.T) {
		file, err := NewMediaFile("testdata/insta360.insp")
		require.NoError(t, err)
		assert.Nil(t, FindInsta360Capture(file))
	})
}

// TestInsta360SkipConvert verifies that only the right lens and proxy of a video capture are skipped.
func TestInsta360SkipConvert(t *testing.T) {
	dir := t.TempDir()
	names := map[string]bool{
		"VID_20220625_140410_00_008.insv": false,
		"VID_20220625_140410_10_008.insv": true,
		"LRV_20220625_140410_11_008.insv": true,
		"IMG_20220625_140410_00_008.insp": false,
		"IMG_20220625_140410_10_008.insp": false,
	}

	for name := range names {
		writeInsta360CaptureFile(t, dir, name, "testdata/flash.jpg")
	}

	for name, skip := range names {
		f, err := NewMediaFile(filepath.Join(dir, name))
		require.NoError(t, err)
		assert.Equal(t, skip, insta360SkipConvert(f), name)
	}

	single, err := NewMediaFile("testdata/insta360.insp")
	require.NoError(t, err)
	assert.False(t, insta360SkipConvert(single))
	assert.False(t, insta360SkipConvert(nil))

	t.Run("Proxy", func(t *testing.T) {
		proxyDir := t.TempDir()
		proxy, proxyErr := NewMediaFile(writeInsta360CaptureFile(t, proxyDir, "LRV_20240415_213145_01_035.lrv", "testdata/flash.jpg"))
		require.NoError(t, proxyErr)
		assert.True(t, insta360SkipConvert(proxy))

		left, leftErr := NewMediaFile(writeInsta360CaptureFile(t, proxyDir, "VID_20240415_213145_00_035.insv", "testdata/flash.jpg"))
		require.NoError(t, leftErr)
		assert.True(t, insta360SkipConvert(proxy))
		assert.False(t, insta360SkipConvert(left))

		other, otherErr := NewMediaFile(writeInsta360CaptureFile(t, proxyDir, "GL010123.LRV", "testdata/flash.jpg"))
		require.NoError(t, otherErr)
		assert.True(t, insta360SkipConvert(other))
	})
}

// TestInsta360ProxyPartner verifies that a left lens and its LRV proxy find each other by name.
func TestInsta360ProxyPartner(t *testing.T) {
	dir := t.TempDir()
	leftName := writeInsta360CaptureFile(t, dir, "VID_20240415_213145_00_035.insv", "testdata/flash.jpg")
	proxyName := writeInsta360CaptureFile(t, dir, "LRV_20240415_213145_01_035.lrv", "testdata/flash.jpg")

	left, err := NewMediaFile(leftName)
	require.NoError(t, err)
	proxy, err := NewMediaFile(proxyName)
	require.NoError(t, err)

	assert.Equal(t, proxyName, insta360ProxyPartner(left))
	assert.Equal(t, leftName, insta360ProxyPartner(proxy))

	for _, name := range []string{
		"VID_20240415_213145_10_035.insv",
		"LRV_20240415_213145_11_035.insv",
		"lrv_20240415_213145_01_036.lrv",
		"LRV_20240415_213145_01_037.LRV",
		"LRV_20240415_213145_01_038.lrv",
		"vid_20240415_213145_00_039.insv",
	} {
		f, fileErr := NewMediaFile(writeInsta360CaptureFile(t, dir, name, "testdata/flash.jpg"))
		require.NoError(t, fileErr)
		assert.Equal(t, "", insta360ProxyPartner(f), name)
	}

	assert.Equal(t, "", insta360ProxyPartner(nil))
}

// TestForceDewarpPreview verifies that only forced runs replace recognized 360° previews.
func TestForceDewarpPreview(t *testing.T) {
	dir := t.TempDir()
	leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
	writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")

	left, err := NewMediaFile(leftName)
	if err != nil {
		t.Fatal(err)
	}

	ordinary, err := NewMediaFile("testdata/flash.jpg")
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, forceDewarpPreview(left, true))
	assert.False(t, forceDewarpPreview(left, false))
	assert.False(t, forceDewarpPreview(ordinary, true))
	assert.False(t, forceDewarpPreview(nil, true))
}

// newInsta360PreviewFixture writes a complete capture below the originals path and returns its left
// lens and the sidecar directory for its previews.
func newInsta360PreviewFixture(t *testing.T, dir string) (*MediaFile, string) {
	t.Helper()
	conf := config.TestConfig()
	originals := filepath.Join(conf.OriginalsPath(), dir)
	sidecars := filepath.Join(conf.SidecarPath(), dir)
	t.Cleanup(func() {
		_ = os.RemoveAll(originals)
		_ = os.RemoveAll(sidecars)
	})

	left, err := NewMediaFile(writeInsta360CaptureFile(t, originals, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg"))
	require.NoError(t, err)
	writeInsta360CaptureFile(t, originals, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
	writeInsta360CaptureFile(t, originals, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg")

	return left, sidecars
}

// TestInsta360PairPreview verifies that only the preview of a complete capture's left lens is recognized.
func TestInsta360PairPreview(t *testing.T) {
	left, sidecars := newInsta360PreviewFixture(t, "pair-preview")

	for name, expected := range map[string]bool{
		"VID_20220625_140410_00_008.insv.jpg": true,
		"VID_20220625_140410_10_008.insv.jpg": false,
		"LRV_20220625_140410_11_008.insv.jpg": false,
		"VID_20220625_140410_00_008.insv.avc": false,
	} {
		preview, err := NewMediaFile(writeInsta360CaptureFile(t, sidecars, name, "testdata/insta360.insp.jpg"))
		require.NoError(t, err)
		capture := insta360PairPreview(preview)
		assert.Equal(t, expected, capture != nil, name)
		if capture != nil {
			assert.Equal(t, left.FileName(), capture.Left.FileName())
		}
	}

	t.Run("IncompleteCapture", func(t *testing.T) {
		_, sidecars := newInsta360PreviewFixture(t, "pair-preview-incomplete")
		require.NoError(t, os.Remove(filepath.Join(Config().OriginalsPath(), "pair-preview-incomplete", "VID_20220625_140410_10_008.insv")))
		preview, err := NewMediaFile(writeInsta360CaptureFile(t, sidecars, "VID_20220625_140410_00_008.insv.jpg", "testdata/insta360.insp.jpg"))
		require.NoError(t, err)
		assert.Nil(t, insta360PairPreview(preview))
	})
	t.Run("Ordinary", func(t *testing.T) {
		ordinary, err := NewMediaFile("testdata/flash.jpg")
		require.NoError(t, err)
		assert.Nil(t, insta360PairPreview(ordinary))
		assert.Nil(t, insta360PairPreview(nil))
	})
}

// TestInsta360RightLensSidecar verifies that only sidecars of a complete capture's right lens are recognized.
func TestInsta360RightLensSidecar(t *testing.T) {
	_, sidecars := newInsta360PreviewFixture(t, "right-lens-sidecar")

	for name, expected := range map[string]bool{
		"VID_20220625_140410_00_008.insv.jpg": false,
		"VID_20220625_140410_10_008.insv.jpg": true,
		"VID_20220625_140410_10_008.insv.avc": true,
		"LRV_20220625_140410_11_008.insv.jpg": false,
	} {
		sidecar, err := NewMediaFile(writeInsta360CaptureFile(t, sidecars, name, "testdata/flash.jpg"))
		require.NoError(t, err)
		assert.Equal(t, expected, insta360RightLensSidecar(sidecar), name)
	}

	ordinary, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)
	assert.False(t, insta360RightLensSidecar(ordinary))
	assert.False(t, insta360RightLensSidecar(nil))
}

// TestInsta360StalePreview verifies that a left lens preview is stale while the right lens is pending,
// unless the preview has a 2:1 aspect ratio.
func TestInsta360StalePreview(t *testing.T) {
	rightLens := func(t *testing.T, dir string) MediaFiles {
		right, err := NewMediaFile(filepath.Join(Config().OriginalsPath(), dir, "VID_20220625_140410_10_008.insv"))
		require.NoError(t, err)
		return MediaFiles{right}
	}

	t.Run("Square", func(t *testing.T) {
		left, sidecars := newInsta360PreviewFixture(t, "stale-preview-square")
		writeInsta360CaptureFile(t, sidecars, "VID_20220625_140410_00_008.insv.jpg", "testdata/flash.jpg")
		assert.True(t, insta360StalePreview(left, rightLens(t, "stale-preview-square")))
	})
	t.Run("RightLensIndexed", func(t *testing.T) {
		left, sidecars := newInsta360PreviewFixture(t, "stale-preview-indexed")
		writeInsta360CaptureFile(t, sidecars, "VID_20220625_140410_00_008.insv.jpg", "testdata/flash.jpg")
		assert.False(t, insta360StalePreview(left, MediaFiles{left}))
		assert.False(t, insta360StalePreview(left, nil))
	})
	t.Run("Combined", func(t *testing.T) {
		left, sidecars := newInsta360PreviewFixture(t, "stale-preview-combined")
		writeInsta360CaptureFile(t, sidecars, "VID_20220625_140410_00_008.insv.jpg", "testdata/insta360.insp.jpg")
		assert.False(t, insta360StalePreview(left, rightLens(t, "stale-preview-combined")))
	})
	t.Run("NoPreview", func(t *testing.T) {
		left, _ := newInsta360PreviewFixture(t, "stale-preview-none")
		assert.False(t, insta360StalePreview(left, rightLens(t, "stale-preview-none")))
	})
	t.Run("RightLens", func(t *testing.T) {
		_, sidecars := newInsta360PreviewFixture(t, "stale-preview-right")
		writeInsta360CaptureFile(t, sidecars, "VID_20220625_140410_10_008.insv.jpg", "testdata/flash.jpg")
		pending := rightLens(t, "stale-preview-right")
		assert.False(t, insta360StalePreview(pending[0], pending))
	})
	t.Run("Ordinary", func(t *testing.T) {
		ordinary, err := NewMediaFile("testdata/flash.jpg")
		require.NoError(t, err)
		assert.False(t, insta360StalePreview(ordinary, MediaFiles{ordinary}))
		assert.False(t, insta360StalePreview(nil, nil))
	})
}

// TestInsta360Capture_MemberPreview verifies that only previews of the right lens and proxy match.
func TestInsta360Capture_MemberPreview(t *testing.T) {
	left, _ := newInsta360PreviewFixture(t, "member-preview")
	capture := FindInsta360Capture(left)
	require.NotNil(t, capture)

	assert.True(t, capture.MemberPreview("member-preview/VID_20220625_140410_10_008.insv.jpg"))
	assert.True(t, capture.MemberPreview("member-preview/LRV_20220625_140410_11_008.insv.jpg"))
	assert.False(t, capture.MemberPreview("member-preview/VID_20220625_140410_00_008.insv.jpg"))
	assert.False(t, capture.MemberPreview("other/VID_20220625_140410_10_008.insv.jpg"))
	assert.False(t, capture.MemberPreview(""))
	assert.False(t, (*Insta360Capture)(nil).MemberPreview("member-preview/VID_20220625_140410_10_008.insv.jpg"))
}

// TestDewarpedVideoFile verifies direct AVC selection, LRV fallback, and fail-closed behavior.
func TestDewarpedVideoFile(t *testing.T) {
	conf := config.TestConfig()
	dir := t.TempDir()
	leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
	writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
	proxyName := writeInsta360CaptureFile(t, dir, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg")

	left, err := NewMediaFile(leftName)
	require.NoError(t, err)
	proxy, err := NewMediaFile(proxyName)
	require.NoError(t, err)
	assert.Nil(t, DewarpedVideoFile(left))

	proxyAvcName, err := fs.FileName(proxy.FileName(), conf.SidecarPath(), conf.OriginalsPath(), fs.ExtAvc)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(proxyAvcName) })
	videoFixture, err := NewMediaFile(conf.SamplesPath() + "/blue-go-video.mp4")
	require.NoError(t, err)
	require.NoError(t, videoFixture.Copy(proxyAvcName, false))
	proxyAvc := DewarpedVideoFile(left)
	require.NotNil(t, proxyAvc)
	assert.Equal(t, proxyAvcName, proxyAvc.FileName())

	leftAvcName, err := fs.FileName(left.FileName(), conf.SidecarPath(), conf.OriginalsPath(), fs.ExtAvc)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(leftAvcName) })
	require.NoError(t, videoFixture.Copy(leftAvcName, false))
	leftAvc := DewarpedVideoFile(left)
	require.NotNil(t, leftAvc)
	assert.Equal(t, leftAvcName, leftAvc.FileName())

	ordinary, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)
	assert.Nil(t, DewarpedVideoFile(ordinary))
	assert.Nil(t, DewarpedVideoFile(nil))
}

// TestInsta360Capture_ValidPair verifies geometry and timing safeguards.
func TestInsta360Capture_ValidPair(t *testing.T) {
	dir := t.TempDir()
	left, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg"))
	require.NoError(t, err)
	right, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg"))
	require.NoError(t, err)
	left.width, left.height = 3072, 3072
	right.width, right.height = 3072, 3072
	capture := &Insta360Capture{Left: left, Right: right}

	assert.True(t, capture.ValidPair())
	invalidRight, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_009.insv", "testdata/2015-02-04.jpg"))
	require.NoError(t, err)
	invalidRight.width, invalidRight.height = 1920, 1080
	assert.False(t, (&Insta360Capture{Left: left, Right: invalidRight}).ValidPair())
	assert.False(t, (*Insta360Capture)(nil).ValidPair())

	// Only video lens files can form a pair.
	photo, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg"))
	require.NoError(t, err)
	photo.width, photo.height = 3072, 3072
	assert.False(t, (&Insta360Capture{Left: left, Right: photo}).ValidPair())
}

// TestAbsDuration verifies duration normalization.
func TestAbsDuration(t *testing.T) {
	assert.Equal(t, absDuration(-5), absDuration(5))
}

// newInsta360StreamFile copies the two-stream video fixture to name below dir and returns it with
// unknown dimensions, as an .insv without readable metadata.
func newInsta360StreamFile(t *testing.T, dir, name string) *MediaFile {
	t.Helper()
	require.NoError(t, fs.MkdirAll(dir))
	fileName := filepath.Join(dir, name)
	require.NoError(t, fs.Copy("../../pkg/media/video/testdata/two-stream.mp4", fileName, false))

	m, err := NewMediaFile(fileName)
	require.NoError(t, err)
	m.width, m.height = 0, 0

	return m
}

// TestMediaFile_Insta360DualStream verifies that only an .insv with two equal square streams matches.
func TestMediaFile_Insta360DualStream(t *testing.T) {
	dir := t.TempDir()

	assert.True(t, newInsta360StreamFile(t, dir, "dual.insv").Insta360DualStream())
	assert.False(t, newInsta360StreamFile(t, dir, "dual.mp4").Insta360DualStream())

	single := newInsta360StreamFile(t, dir, "single.insv")
	single.videoOnce.Do(func() {})
	single.videoInfo.TrackSizes = []video.TrackSize{{Width: 64, Height: 64}}
	assert.False(t, single.Insta360DualStream())

	unequal := newInsta360StreamFile(t, dir, "unequal.insv")
	unequal.videoOnce.Do(func() {})
	unequal.videoInfo.TrackSizes = []video.TrackSize{{Width: 64, Height: 64}, {Width: 32, Height: 32}}
	assert.False(t, unequal.Insta360DualStream())

	wide := newInsta360StreamFile(t, dir, "wide.insv")
	wide.videoOnce.Do(func() {})
	wide.videoInfo.TrackSizes = []video.TrackSize{{Width: 128, Height: 64}, {Width: 128, Height: 64}}
	assert.False(t, wide.Insta360DualStream())

	assert.False(t, (*MediaFile)(nil).Insta360DualStream())
}

// TestMediaFile_DualFisheyeLayoutTrackSize verifies that an .insv without known dimensions uses the
// size of its first video track.
func TestMediaFile_DualFisheyeLayoutTrackSize(t *testing.T) {
	dir := t.TempDir()

	assert.False(t, newInsta360StreamFile(t, dir, "square.insv").DualFisheyeLayout())

	wide := newInsta360StreamFile(t, dir, "wide.insv")
	wide.videoOnce.Do(func() {})
	wide.videoInfo.TrackSizes = []video.TrackSize{{Width: 768, Height: 384}}
	assert.True(t, wide.DualFisheyeLayout())

	unknown := newInsta360StreamFile(t, dir, "unknown.insv")
	unknown.videoOnce.Do(func() {})
	assert.True(t, unknown.DualFisheyeLayout())
}

// TestInsta360ExpectsDewarp verifies which Insta360 originals are expected to get a dewarped preview.
func TestInsta360ExpectsDewarp(t *testing.T) {
	dir := t.TempDir()

	assert.True(t, insta360ExpectsDewarp(newInsta360StreamFile(t, dir, "dual.insv")))

	single := newInsta360StreamFile(t, dir, "single.insv")
	single.videoOnce.Do(func() {})
	single.videoInfo.TrackSizes = []video.TrackSize{{Width: 64, Height: 64}}
	assert.False(t, insta360ExpectsDewarp(single))

	insp, err := NewMediaFile("testdata/insta360.insp")
	require.NoError(t, err)
	assert.True(t, insta360ExpectsDewarp(insp))

	ordinary, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)
	assert.False(t, insta360ExpectsDewarp(ordinary))
	assert.False(t, insta360ExpectsDewarp(nil))
}
