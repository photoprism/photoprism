package photoprism

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
)

const (
	takenCapture   = "2020:10:26 15:46:29"
	takenWriteTime = "2020:10:26 15:46:31"
)

// writeTakenFile writes a JPEG image under the specified name with the specified ExifTool tag arguments.
func writeTakenFile(t *testing.T, cfg *config.Config, fileName string, tags ...string) {
	t.Helper()

	if !cfg.ExifToolEnabled() {
		t.Skip("ExifTool must be available to write time tags")
	}

	writeInsta360Photo(t, cfg, fileName, "320x160")

	// The description keeps files with the same tags from being skipped as duplicates.
	args := append([]string{"-q", "-overwrite_original", "-all=", "-EXIF:ImageDescription=" + filepath.Base(fileName)}, tags...)
	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.ExifToolBin(), append(args, fileName)...).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))
}

// takenAtOf returns the time the photo owning the specified original was taken, as stored in the index.
func takenAtOf(t *testing.T, relName string) entity.Photo {
	t.Helper()

	var file entity.File
	require.NoError(t, entity.UnscopedDb().First(&file, "file_root = ? AND file_name = ?", entity.RootOriginals, relName).Error)

	var photo entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", file.PhotoID).Error)

	return photo
}

// TestIndex_TakenAtModifyTime verifies that a modify time never replaces the capture time of another file of the
// same stack, whatever order the files are indexed in, and ranks below a date from the file name.
func TestIndex_TakenAtModifyTime(t *testing.T) {
	const (
		capture   = "IMG_20201026_154628_00_070.insp"
		writeTime = "IMG_20201026_154628_00_070.jpg"
	)

	expected := time.Date(2020, 10, 26, 15, 46, 29, 0, time.UTC)

	for _, tc := range []struct {
		name  string
		first []string
		late  []string
	}{
		{"SameRun", []string{capture, writeTime}, nil},
		{"WriteTimeFirst", []string{writeTime}, []string{capture}},
		{"CaptureFirst", []string{capture}, []string{writeTime}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			folder := strings.ToLower("takenwritetime" + tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			write := func(name string) {
				if name == capture {
					writeTakenFile(t, cfg, filepath.Join(dir, name), "-EXIF:DateTimeOriginal="+takenCapture, "-EXIF:ModifyDate="+takenCapture)
				} else {
					writeTakenFile(t, cfg, filepath.Join(dir, name), "-EXIF:ModifyDate="+takenWriteTime)
				}
			}

			for _, name := range tc.first {
				write(name)
			}

			indexInsta360StackFolder(cfg, folder, false, true)

			for _, name := range tc.late {
				write(name)
			}

			if len(tc.late) > 0 {
				indexInsta360StackFolder(cfg, folder, false, true)
			}

			photo := takenAtOf(t, folder+"/"+capture)
			assert.Equal(t, photo.ID, takenAtOf(t, folder+"/"+writeTime).ID)
			assert.Equal(t, expected, photo.TakenAt.UTC())
			assert.Equal(t, entity.SrcMeta, photo.TakenSrc)

			// A forced rescan keeps the capture time.
			indexInsta360StackFolder(cfg, folder, true, false)
			assert.Equal(t, expected, takenAtOf(t, folder+"/"+capture).TakenAt.UTC())
		})
	}
	t.Run("ModifyTimeOnly", func(t *testing.T) {
		folder := "takenwritetimeonly"
		cfg := newInsta360StackConfig(t, folder, false)
		writeTakenFile(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, "IMG_2020-10-26_15-46-28.jpg"), "-EXIF:ModifyDate="+takenWriteTime)
		indexInsta360StackFolder(cfg, folder, false, true)

		// The date in the file name, which is three seconds earlier, ranks above the modify time.
		photo := takenAtOf(t, folder+"/IMG_2020-10-26_15-46-28.jpg")
		assert.Equal(t, time.Date(2020, 10, 26, 15, 46, 28, 0, time.UTC), photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcName, photo.TakenSrc)
	})
	t.Run("ModifyTimeWithoutName", func(t *testing.T) {
		folder := "takenwritetimenoname"
		cfg := newInsta360StackConfig(t, folder, false)
		writeTakenFile(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, "write-time.jpg"), "-EXIF:ModifyDate="+takenWriteTime)
		indexInsta360StackFolder(cfg, folder, false, true)

		photo := takenAtOf(t, folder+"/write-time.jpg")
		assert.Equal(t, time.Date(2020, 10, 26, 15, 46, 31, 0, time.UTC), photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcModified, photo.TakenSrc)
	})
	t.Run("NonPrimaryModifyTime", func(t *testing.T) {
		// The modify time of a JPEG that is not primary is never used.
		folder := "takenwritetimenonprimarymod"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		writeTakenFile(t, cfg, filepath.Join(dir, "stack.jpg"))
		indexInsta360StackFolder(cfg, folder, false, true)
		require.Equal(t, entity.SrcAuto, takenAtOf(t, folder+"/stack.jpg").TakenSrc)

		writeTakenFile(t, cfg, filepath.Join(dir, "stack.jpeg"), "-EXIF:ModifyDate="+takenWriteTime)
		indexInsta360StackFolder(cfg, folder, false, true)

		photo := takenAtOf(t, folder+"/stack.jpeg")
		assert.Equal(t, photo.ID, takenAtOf(t, folder+"/stack.jpg").ID)
		assert.Equal(t, entity.SrcAuto, photo.TakenSrc)
	})
	t.Run("NonPrimaryCapture", func(t *testing.T) {
		// The capture time of a JPEG that is not primary replaces the modify time of the primary JPEG.
		folder := "takenwritetimenonprimary"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		writeTakenFile(t, cfg, filepath.Join(dir, "stack.jpg"), "-EXIF:ModifyDate="+takenWriteTime)
		indexInsta360StackFolder(cfg, folder, false, true)
		require.Equal(t, entity.SrcModified, takenAtOf(t, folder+"/stack.jpg").TakenSrc)

		writeTakenFile(t, cfg, filepath.Join(dir, "stack.jpeg"), "-EXIF:DateTimeOriginal="+takenCapture, "-EXIF:OffsetTimeOriginal=+02:00")
		indexInsta360StackFolder(cfg, folder, false, true)

		// The time zone from the metadata is kept, as when the capture time is indexed first.
		photo := takenAtOf(t, folder+"/stack.jpeg")
		assert.Equal(t, photo.ID, takenAtOf(t, folder+"/stack.jpg").ID)
		assert.Equal(t, time.Date(2020, 10, 26, 13, 46, 29, 0, time.UTC), photo.TakenAt.UTC())
		assert.Equal(t, expected, photo.TakenAtLocal.UTC())
		assert.Equal(t, "UTC+2", photo.TimeZone)
		assert.Equal(t, entity.SrcMeta, photo.TakenSrc)
	})
}

// TestSetTakenAtMeta verifies that a modify time ranks below capture times and file names, and keeps the earliest.
func TestSetTakenAtMeta(t *testing.T) {
	taken := time.Date(2020, 10, 26, 15, 46, 29, 0, time.UTC)
	modified := time.Date(2020, 10, 26, 15, 46, 31, 0, time.UTC)

	t.Run("CaptureTimeFirst", func(t *testing.T) {
		photo := entity.NewPhoto(true)
		setTakenAtMeta(&photo, meta.Data{TakenAt: taken, TakenAtLocal: taken})
		setTakenAtMeta(&photo, meta.Data{ModifiedAt: modified})
		assert.Equal(t, taken, photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcMeta, photo.TakenSrc)
	})
	t.Run("ModifyTimeFirst", func(t *testing.T) {
		photo := entity.NewPhoto(true)
		setTakenAtMeta(&photo, meta.Data{ModifiedAt: modified})
		assert.Equal(t, modified, photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcModified, photo.TakenSrc)
		setTakenAtMeta(&photo, meta.Data{TakenAt: taken, TakenAtLocal: taken})
		assert.Equal(t, taken, photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcMeta, photo.TakenSrc)
	})
	t.Run("BelowFileName", func(t *testing.T) {
		photo := entity.NewPhoto(true)
		photo.SetTakenAt(taken, taken, "", entity.SrcName)
		setTakenAtMeta(&photo, meta.Data{ModifiedAt: modified})
		assert.Equal(t, taken, photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcName, photo.TakenSrc)
	})
	t.Run("EarliestModifyTime", func(t *testing.T) {
		photo := entity.NewPhoto(true)
		setTakenAtMeta(&photo, meta.Data{ModifiedAt: modified})
		setTakenAtMeta(&photo, meta.Data{ModifiedAt: modified.Add(time.Hour)})
		assert.Equal(t, modified, photo.TakenAt.UTC())
		setTakenAtMeta(&photo, meta.Data{ModifiedAt: taken})
		assert.Equal(t, taken, photo.TakenAt.UTC())
		assert.Equal(t, entity.SrcModified, photo.TakenSrc)
	})
}

// TestMediaTimeUTC verifies that the modify time is only used without a capture time.
func TestMediaTimeUTC(t *testing.T) {
	taken := time.Date(2020, 10, 26, 15, 46, 29, 0, time.UTC)
	modified := time.Date(2020, 10, 26, 15, 46, 31, 0, time.UTC)

	assert.Equal(t, taken, mediaTimeUTC(meta.Data{TakenAt: taken, ModifiedAt: modified}))
	assert.Equal(t, modified, mediaTimeUTC(meta.Data{ModifiedAt: modified}))
	assert.True(t, mediaTimeUTC(meta.Data{}).IsZero())
}

// TestIndex_StackMetaModifyTime verifies that pictures are stacked by capture time and place, never by modify time.
func TestIndex_StackMetaModifyTime(t *testing.T) {
	gps := []string{"-GPSLatitude=52.52", "-GPSLatitudeRef=N", "-GPSLongitude=13.405", "-GPSLongitudeRef=E"}

	for _, tc := range []struct {
		name    string
		time    string
		stacked bool
	}{
		{"ModifyTime", "-EXIF:ModifyDate=" + takenWriteTime, false},
		{"CaptureTime", "-EXIF:DateTimeOriginal=" + takenCapture, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			folder := strings.ToLower("stackmeta" + tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			cfg.Settings().Stack.Meta = true
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			for _, name := range []string{"first.jpg", "second.jpg"} {
				writeTakenFile(t, cfg, filepath.Join(dir, name), append([]string{tc.time}, gps...)...)
				indexInsta360StackFolder(cfg, folder, false, true)
			}

			assert.Equal(t, tc.stacked, takenAtOf(t, folder+"/first.jpg").ID == takenAtOf(t, folder+"/second.jpg").ID)
		})
	}
}
