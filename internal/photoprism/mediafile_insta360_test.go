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
		dir := t.TempDir()
		writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg")
		rightName := writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg")

		right, err := NewMediaFile(rightName)
		require.NoError(t, err)
		capture := FindInsta360Capture(right)
		require.NotNil(t, capture)
		assert.False(t, capture.Video())
		assert.True(t, capture.ValidPair())
		assert.Nil(t, capture.Proxy)
		assert.Len(t, capture.Files(), 2)
		assert.Equal(t, media.Insta360VideoRight, capture.Name.Role)
	})
	t.Run("PhotoSingleFileCaptureName", func(t *testing.T) {
		dir := t.TempDir()
		fileName := writeInsta360CaptureFile(t, dir, "IMG_20231015_101112_00_123.insp", "testdata/insta360.insp")
		file, err := NewMediaFile(fileName)
		require.NoError(t, err)

		capture := FindInsta360Capture(file)
		require.NotNil(t, capture)
		assert.Equal(t, fileName, capture.Left.FileName())
		assert.Nil(t, capture.Right)
		assert.False(t, capture.ValidPair())

		related, err := file.RelatedFiles(false)
		require.NoError(t, err)
		assert.Len(t, related.Files, 1)
		assert.Equal(t, fileName, related.Main.FileName())
	})
	t.Run("CaseVariant", func(t *testing.T) {
		dir := t.TempDir()
		writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg")
		variant, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "img_20220625_140410_00_008.insp", "testdata/flash.jpg"))
		require.NoError(t, err)
		upperExt, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.INSV", "testdata/flash.jpg"))
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

// TestInsta360Capture_ValidVideoPair verifies that only complete video captures qualify.
func TestInsta360Capture_ValidVideoPair(t *testing.T) {
	dir := t.TempDir()
	videoLeft, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg"))
	require.NoError(t, err)
	writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
	photoLeft, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg"))
	require.NoError(t, err)
	writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg")

	assert.True(t, FindInsta360Capture(videoLeft).ValidVideoPair())
	assert.True(t, FindInsta360Capture(photoLeft).ValidPair())
	assert.False(t, FindInsta360Capture(photoLeft).ValidVideoPair())
	assert.False(t, (*Insta360Capture)(nil).ValidVideoPair())
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
}

// TestInsta360Capture_Video verifies that video and photo captures are told apart.
func TestInsta360Capture_Video(t *testing.T) {
	assert.True(t, (&Insta360Capture{}).Video())
	assert.False(t, (&Insta360Capture{Name: media.Insta360VideoName{Photo: true}}).Video())
	assert.False(t, (*Insta360Capture)(nil).Video())
}

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

	// Photo pairs are compared by size only, and both files must be photos.
	photoLeft, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg"))
	require.NoError(t, err)
	photoRight, err := NewMediaFile(writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg"))
	require.NoError(t, err)
	photoLeft.width, photoLeft.height = 3072, 3072
	photoRight.width, photoRight.height = 3072, 3072
	photoName := media.Insta360VideoName{Photo: true}
	assert.True(t, (&Insta360Capture{Name: photoName, Left: photoLeft, Right: photoRight}).ValidPair())
	assert.False(t, (&Insta360Capture{Name: photoName, Left: photoLeft, Right: right}).ValidPair())
	assert.False(t, (&Insta360Capture{Left: left, Right: photoRight}).ValidPair())
	photoRight.width, photoRight.height = 1024, 1024
	assert.False(t, (&Insta360Capture{Name: photoName, Left: photoLeft, Right: photoRight}).ValidPair())
}

// TestAbsDuration verifies duration normalization.
func TestAbsDuration(t *testing.T) {
	assert.Equal(t, absDuration(-5), absDuration(5))
}
