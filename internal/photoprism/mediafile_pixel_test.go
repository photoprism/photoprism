package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// writeGooglePixelCaptureFile copies a fixture to a Google Pixel Camera capture filename for unit tests.
func writeGooglePixelCaptureFile(t *testing.T, dir, name, fixture string) string {
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

// TestFindGooglePixelCapture verifies multi-file capture grouping, primary selection, and single-file fallback.
func TestFindGooglePixelCapture(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")
	videoFixture := filepath.Join(cfg.SamplesPath(), "blue-go-video.mp4")

	t.Run("StandardPhotoRaw", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
		rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)
		rawMedia, err := NewMediaFile(rawName)
		require.NoError(t, err)

		captureFromCover := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, captureFromCover)
		assert.True(t, captureFromCover.ValidPair())
		assert.Equal(t, coverName, captureFromCover.Primary().FileName())
		assert.Len(t, captureFromCover.Files(), 2)

		captureFromRaw := FindGooglePixelCapture(rawMedia)
		require.NotNil(t, captureFromRaw)
		assert.True(t, captureFromRaw.ValidPair())
		assert.Equal(t, coverName, captureFromRaw.Primary().FileName())
		assert.Len(t, captureFromRaw.Files(), 2)
	})

	t.Run("LegacyCover", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.COVER.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
	})

	t.Run("MotionPhoto", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.MP.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
	})

	t.Run("NightSight", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.NIGHT.RAW-01.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.NIGHT.RAW-02.ORIGINAL.dng", rawFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
	})

	t.Run("PortraitModern", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_150000123.PORTRAIT.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_150000123.PORTRAIT.ORIGINAL.jpg", jpegFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
		assert.Len(t, capture.Files(), 2)
	})

	t.Run("PortraitLegacy", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_150000123.PORTRAIT-01.COVER.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_150000123.PORTRAIT-02.ORIGINAL.jpg", jpegFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
	})

	t.Run("Burst", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_160000123.BURST-01.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_160000123.BURST-02.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_160000123.BURST-03.jpg", jpegFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
		assert.Len(t, capture.Files(), 3)
	})

	t.Run("LongExposure", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_170000123.LONG_EXPOSURE-01.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_170000123.LONG_EXPOSURE-02.ORIGINAL.jpg", jpegFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
	})

	t.Run("ActionPan", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_170000123.ACTION_PAN-01.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240926_170000123.ACTION_PAN-02.ORIGINAL.jpg", jpegFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
	})

	t.Run("AIProZoom", func(t *testing.T) {
		dir := t.TempDir()
		coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_120000123.BURST-01.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240930_120000123.BURST-02.original.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "PXL_20240930_120000123.BURST-03.ORIGINAL.dng", rawFixture)

		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, coverName, capture.Primary().FileName())
		assert.Len(t, capture.Files(), 3)
	})

	t.Run("VideoBoost", func(t *testing.T) {
		dir := t.TempDir()
		coverVideo := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_180000123.VB-01.COVER.mp4", videoFixture)
		mainVideo := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_180000123.VB-03.MAIN.mp4", videoFixture)

		coverMedia, err := NewMediaFile(coverVideo)
		require.NoError(t, err)
		mainMedia, err := NewMediaFile(mainVideo)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(coverMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, mainVideo, capture.Primary().FileName())
		assert.Len(t, capture.Files(), 2)

		captureFromMain := FindGooglePixelCapture(mainMedia)
		require.NotNil(t, captureFromMain)
		assert.True(t, captureFromMain.ValidPair())
		assert.Equal(t, mainVideo, captureFromMain.Primary().FileName())
	})

	t.Run("NightSightVideo", func(t *testing.T) {
		dir := t.TempDir()
		writeGooglePixelCaptureFile(t, dir, "PXL_20260930_200000123.NS-01.COVER.mp4", videoFixture)
		mainVideo := writeGooglePixelCaptureFile(t, dir, "PXL_20260930_200000123.NS-02.MAIN.mp4", videoFixture)

		mainMedia, err := NewMediaFile(mainVideo)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(mainMedia)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Equal(t, mainVideo, capture.Primary().FileName())
	})

	t.Run("MultipleMainPhotosPreserved", func(t *testing.T) {
		dir := t.TempDir()
		photo1 := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
		photo2 := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.COVER.jpg", jpegFixture)

		f1, err := NewMediaFile(photo1)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(f1)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Len(t, capture.Files(), 2, "displaced MainPhoto must be preserved in Originals")
		assert.True(t, capture.IsMember(photo1))
		assert.True(t, capture.IsMember(photo2))
	})

	t.Run("MultipleDraftVideosPreserved", func(t *testing.T) {
		dir := t.TempDir()
		draft1 := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_180000123.VB-01.COVER.mp4", videoFixture)
		draft2 := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_180000123.VB-01.mp4", videoFixture)
		mainVid := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_180000123.VB-03.MAIN.mp4", videoFixture)

		f, err := NewMediaFile(mainVid)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(f)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
		assert.Len(t, capture.Files(), 3, "displaced draft preview video must be preserved in Originals")
		assert.True(t, capture.IsMember(draft1))
		assert.True(t, capture.IsMember(draft2))
		assert.True(t, capture.IsMember(mainVid))
	})

	t.Run("CaseVariants", func(t *testing.T) {
		dir := t.TempDir()
		cover := writeGooglePixelCaptureFile(t, dir, "pxl_20240926_143000123.raw-01.jpg", jpegFixture)
		writeGooglePixelCaptureFile(t, dir, "pxl_20240926_143000123.raw-02.original.dng", rawFixture)

		f, err := NewMediaFile(cover)
		require.NoError(t, err)

		capture := FindGooglePixelCapture(f)
		require.NotNil(t, capture)
		assert.True(t, capture.ValidPair())
	})

	t.Run("LoneFiles", func(t *testing.T) {
		dirPhoto := t.TempDir()
		lonePhoto := writeGooglePixelCaptureFile(t, dirPhoto, "PXL_20240926_150000123.PORTRAIT.jpg", jpegFixture)
		fPhoto, err := NewMediaFile(lonePhoto)
		require.NoError(t, err)

		capturePhoto := FindGooglePixelCapture(fPhoto)
		require.NotNil(t, capturePhoto)
		assert.False(t, capturePhoto.ValidPair(), "lone photo must not form a valid pair")
		assert.Len(t, capturePhoto.Files(), 1)

		dirRaw := t.TempDir()
		loneRaw := writeGooglePixelCaptureFile(t, dirRaw, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)
		fRaw, err := NewMediaFile(loneRaw)
		require.NoError(t, err)

		captureRaw := FindGooglePixelCapture(fRaw)
		require.NotNil(t, captureRaw)
		assert.False(t, captureRaw.ValidPair(), "lone RAW file must not form a valid pair")
		assert.Len(t, captureRaw.Files(), 1)
	})

	t.Run("UnrelatedFiles", func(t *testing.T) {
		dir := t.TempDir()
		unrelated1 := writeGooglePixelCaptureFile(t, dir, "IMG_20231015_101112.jpg", jpegFixture)
		unrelated2 := writeGooglePixelCaptureFile(t, dir, "PXL_20230805_123456.jpg", jpegFixture)
		unrelated3 := writeGooglePixelCaptureFile(t, dir, "PXL_20230805_123456.MP.jpg", jpegFixture)

		for _, name := range []string{unrelated1, unrelated2, unrelated3} {
			f, err := NewMediaFile(name)
			require.NoError(t, err)
			assert.Nil(t, FindGooglePixelCapture(f), name)
		}

		assert.Nil(t, FindGooglePixelCapture(nil))
	})
}

// TestGooglePixelCapture_Primary verifies precedence among boosted video, main photo, and originals.
func TestGooglePixelCapture_Primary(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	videoFixture := filepath.Join(cfg.SamplesPath(), "blue-go-video.mp4")

	dir := t.TempDir()
	photoPath := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	videoPath := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.VB-03.MAIN.mp4", videoFixture)
	origPath := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.jpg", jpegFixture)

	photo, err := NewMediaFile(photoPath)
	require.NoError(t, err)
	vid, err := NewMediaFile(videoPath)
	require.NoError(t, err)
	orig, err := NewMediaFile(origPath)
	require.NoError(t, err)

	// Video takes precedence over photo and originals.
	cVideo := &GooglePixelCapture{MainVideo: vid, MainPhoto: photo, Originals: MediaFiles{orig}}
	assert.Equal(t, vid.FileName(), cVideo.Primary().FileName())

	// Photo takes precedence over originals.
	cPhoto := &GooglePixelCapture{MainPhoto: photo, Originals: MediaFiles{orig}}
	assert.Equal(t, photo.FileName(), cPhoto.Primary().FileName())

	// Originals[0] fallback when no MainVideo or MainPhoto.
	cOrig := &GooglePixelCapture{Originals: MediaFiles{orig}}
	assert.Equal(t, orig.FileName(), cOrig.Primary().FileName())

	// Empty and nil captures return nil.
	assert.Nil(t, (&GooglePixelCapture{}).Primary())
	assert.Nil(t, (*GooglePixelCapture)(nil).Primary())
}

// TestGooglePixelCapture_ValidPair verifies pair validity requirements.
func TestGooglePixelCapture_ValidPair(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")

	dir := t.TempDir()
	p1Path := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	p2Path := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.jpg", jpegFixture)

	f1, err := NewMediaFile(p1Path)
	require.NoError(t, err)
	f2, err := NewMediaFile(p2Path)
	require.NoError(t, err)

	assert.False(t, (*GooglePixelCapture)(nil).ValidPair())
	assert.False(t, (&GooglePixelCapture{}).ValidPair())
	assert.False(t, (&GooglePixelCapture{MainPhoto: f1}).ValidPair())
	assert.False(t, (&GooglePixelCapture{MainVideo: f1}).ValidPair())
	assert.False(t, (&GooglePixelCapture{Originals: MediaFiles{f1}}).ValidPair())

	assert.True(t, (&GooglePixelCapture{MainPhoto: f1, Originals: MediaFiles{f2}}).ValidPair())
	assert.True(t, (&GooglePixelCapture{MainVideo: f1, Originals: MediaFiles{f2}}).ValidPair())
	assert.True(t, (&GooglePixelCapture{MainPhoto: f1, MainVideo: f2}).ValidPair())
	assert.True(t, (&GooglePixelCapture{Originals: MediaFiles{f1, f2}}).ValidPair())
}

// TestGooglePixelCapture_Files verifies listing member files with primary first and deduplicated originals.
func TestGooglePixelCapture_Files(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")

	dir := t.TempDir()
	p1Path := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	p2Path := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.jpg", jpegFixture)

	f1, err := NewMediaFile(p1Path)
	require.NoError(t, err)
	f2, err := NewMediaFile(p2Path)
	require.NoError(t, err)

	assert.Nil(t, (*GooglePixelCapture)(nil).Files())

	capture := &GooglePixelCapture{
		MainPhoto: f1,
		Originals: MediaFiles{f1, f2}, // f1 is duplicated in Originals
	}

	files := capture.Files()
	require.Len(t, files, 2)
	assert.Equal(t, f1.FileName(), files[0].FileName())
	assert.Equal(t, f2.FileName(), files[1].FileName())

	// When both MainVideo and MainPhoto are present, MainVideo is primary and MainPhoto is preserved as a member.
	captureWithBoth := &GooglePixelCapture{
		MainVideo: f2,
		MainPhoto: f1,
	}
	filesBoth := captureWithBoth.Files()
	require.Len(t, filesBoth, 2)
	assert.Equal(t, f2.FileName(), filesBoth[0].FileName())
	assert.Equal(t, f1.FileName(), filesBoth[1].FileName())
}

// TestGooglePixelCapture_IsPrimary verifies primary determination for media files.
func TestGooglePixelCapture_IsPrimary(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")

	dir := t.TempDir()
	coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

	cover, err := NewMediaFile(coverName)
	require.NoError(t, err)
	raw, err := NewMediaFile(rawName)
	require.NoError(t, err)

	capture := FindGooglePixelCapture(cover)
	require.NotNil(t, capture)

	assert.True(t, capture.IsPrimary(cover))
	assert.False(t, capture.IsPrimary(raw))
	assert.False(t, capture.IsPrimary(nil))
	assert.False(t, (*GooglePixelCapture)(nil).IsPrimary(cover))
}

// TestGooglePixelCapture_IsMember verifies matching by absolute path, root-relative name, and base name.
func TestGooglePixelCapture_IsMember(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")

	dir := t.TempDir()
	coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

	cover, err := NewMediaFile(coverName)
	require.NoError(t, err)
	raw, err := NewMediaFile(rawName)
	require.NoError(t, err)

	capture := FindGooglePixelCapture(cover)
	require.NotNil(t, capture)

	assert.True(t, capture.IsMember(cover.FileName()))
	assert.True(t, capture.IsMember(cover.RootRelName()))
	assert.True(t, capture.IsMember(cover.BaseName()))
	assert.True(t, capture.IsMember(raw.FileName()))
	assert.True(t, capture.IsMember(raw.BaseName()))

	assert.False(t, capture.IsMember("unrelated.jpg"))
	assert.False(t, capture.IsMember(""))
	assert.False(t, (*GooglePixelCapture)(nil).IsMember(cover.FileName()))
}

// TestGooglePixelCapture_MemberFile verifies entity.File recognition including sidecar files.
func TestGooglePixelCapture_MemberFile(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")

	dir := t.TempDir()
	coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

	cover, err := NewMediaFile(coverName)
	require.NoError(t, err)

	capture := FindGooglePixelCapture(cover)
	require.NotNil(t, capture)

	assert.True(t, capture.MemberFile(entity.File{FileName: cover.FileName()}))
	assert.True(t, capture.MemberFile(entity.File{FileName: rawName}))
	assert.True(t, capture.MemberFile(entity.File{FileName: "foo.xmp", FileSidecar: true}))
	assert.True(t, capture.MemberFile(entity.File{FileName: "foo.yml", FileRoot: entity.RootSidecar}))

	assert.False(t, capture.MemberFile(entity.File{FileName: "unrelated.jpg"}))
	assert.False(t, (*GooglePixelCapture)(nil).MemberFile(entity.File{FileName: cover.FileName()}))
}

// TestGooglePixelSkipConvert verifies that companion RAW files are skipped for conversion when primary JPEG exists.
func TestGooglePixelSkipConvert(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")
	videoFixture := filepath.Join(cfg.SamplesPath(), "blue-go-video.mp4")

	dir := t.TempDir()
	coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)
	videoName := writeGooglePixelCaptureFile(t, dir, "PXL_20240930_180000123.VB-03.MAIN.mp4", videoFixture)

	cover, err := NewMediaFile(coverName)
	require.NoError(t, err)
	raw, err := NewMediaFile(rawName)
	require.NoError(t, err)
	vid, err := NewMediaFile(videoName)
	require.NoError(t, err)

	// Companion RAW is skipped because primary JPEG exists.
	assert.True(t, googlePixelSkipConvert(raw))

	// Primary JPEG is not skipped (not a RAW file).
	assert.False(t, googlePixelSkipConvert(cover))

	// Video is not skipped.
	assert.False(t, googlePixelSkipConvert(vid))

	// Standalone RAW without JPEG companion is not skipped.
	standaloneDir := t.TempDir()
	standaloneRawName := writeGooglePixelCaptureFile(t, standaloneDir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)
	standaloneRaw, err := NewMediaFile(standaloneRawName)
	require.NoError(t, err)
	assert.False(t, googlePixelSkipConvert(standaloneRaw))

	assert.False(t, googlePixelSkipConvert(nil))
}

// TestGooglePixelImportOrder verifies that the primary file is moved to index 0.
func TestGooglePixelImportOrder(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")

	dir := t.TempDir()
	coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

	cover, err := NewMediaFile(coverName)
	require.NoError(t, err)
	raw, err := NewMediaFile(rawName)
	require.NoError(t, err)

	// Primary cover passed as Main: primary should be at index 0.
	ordered := googlePixelImportOrder(RelatedFiles{Main: cover, Files: MediaFiles{raw, cover}})
	require.Len(t, ordered, 2)
	assert.Equal(t, cover.FileName(), ordered[0].FileName())
	assert.Equal(t, raw.FileName(), ordered[1].FileName())

	// RAW passed as Main (not primary): order untouched.
	unchanged := googlePixelImportOrder(RelatedFiles{Main: raw, Files: MediaFiles{raw, cover}})
	require.Len(t, unchanged, 2)
	assert.Equal(t, raw.FileName(), unchanged[0].FileName())

	// Non-Pixel files: order untouched.
	unrelatedFile, err := NewMediaFile(jpegFixture)
	require.NoError(t, err)
	nonPixel := MediaFiles{unrelatedFile}
	assert.Equal(t, nonPixel, googlePixelImportOrder(RelatedFiles{Main: unrelatedFile, Files: nonPixel}))
}

// TestGooglePixelPrimaryCapture verifies resolution of primary capture from MediaFile.
func TestGooglePixelPrimaryCapture(t *testing.T) {
	cfg := config.TestConfig()
	jpegFixture := filepath.Join(cfg.SamplesPath(), "beach_sand.jpg")
	rawFixture := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")

	dir := t.TempDir()
	coverName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	rawName := writeGooglePixelCaptureFile(t, dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng", rawFixture)

	cover, err := NewMediaFile(coverName)
	require.NoError(t, err)
	raw, err := NewMediaFile(rawName)
	require.NoError(t, err)

	// Primary file returns capture.
	capture := googlePixelPrimaryCapture(cover)
	require.NotNil(t, capture)
	assert.True(t, capture.IsPrimary(cover))

	// Companion RAW does not return capture (it is not primary).
	assert.Nil(t, googlePixelPrimaryCapture(raw))

	// Nil returns nil.
	assert.Nil(t, googlePixelPrimaryCapture(nil))

	// Lone file does not return capture (not a valid multi-file pair).
	loneDir := t.TempDir()
	loneCoverName := writeGooglePixelCaptureFile(t, loneDir, "PXL_20240926_143000123.RAW-01.jpg", jpegFixture)
	loneCover, err := NewMediaFile(loneCoverName)
	require.NoError(t, err)
	assert.Nil(t, googlePixelPrimaryCapture(loneCover))
}
