package photoprism

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

const (
	insta360StackLeft  = "VID_20220625_140410_00_008.insv"
	insta360StackRight = "VID_20220625_140410_10_008.insv"
	insta360StackProxy = "LRV_20220625_140410_11_008.insv"
	insta360StackName  = "VID_20220625_140410_00_008"
)

// newInsta360StackConfig returns an isolated config with its own database and restores the
// previous config when the test ends.
func newInsta360StackConfig(t *testing.T, dbName string, stackSequences bool) *config.Config {
	t.Helper()

	cfg := config.NewMinimalTestConfigWithDb(dbName, filepath.Join(t.TempDir(), "storage"))
	cfg.Settings().Stack.Name = stackSequences

	if !cfg.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to create synthetic capture files")
	}

	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	return cfg
}

// writeInsta360StackMedia writes a one-second square H.264 clip, or a JPEG for image names, with
// bytes that differ per file name.
func writeInsta360StackMedia(t *testing.T, cfg *config.Config, dir, name string) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(dir))

	fileName := filepath.Join(dir, name)
	fileType := fs.FileType(name)
	image := fileType == fs.ImageJpeg || fileType == fs.ImageInsp

	args := []string{"-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x320:rate=30",
		"-t", "1", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-metadata", "title=" + name, "-f", "mp4", fileName}

	if image {
		args = []string{"-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x320",
			"-frames:v", "1", "-f", "image2", "-c:v", "mjpeg", fileName}
	}

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), args...).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))

	// Images get the name appended after the end marker, so files with different names never share a hash.
	if image {
		// #nosec G304 -- the destination directory and filename are controlled by the test.
		f, openErr := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0)
		require.NoError(t, openErr)
		_, writeErr := f.WriteString(name)
		require.NoError(t, writeErr)
		require.NoError(t, f.Close())
	}
}

// indexInsta360StackFolder indexes a folder below the originals path.
func indexInsta360StackFolder(cfg *config.Config, folder string, rescan, skipArchived bool) {
	indexInsta360StackOptions(cfg, NewIndexOptions(folder, rescan, true, true, false, skipArchived, cfg))
}

// indexInsta360StackOptions runs the indexer with the given options.
func indexInsta360StackOptions(cfg *config.Config, opt IndexOptions) {
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	ind.Start(opt)
}

// insta360StackOwners returns the photos that own the originals in folder, keyed by file name.
func insta360StackOwners(t *testing.T, folder string) map[string]entity.Photo {
	t.Helper()

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().
		Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, folder+"/%").
		Find(&files).Error)

	result := make(map[string]entity.Photo, len(files))

	for _, f := range files {
		var p entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&p, "id = ?", f.PhotoID).Error)
		result[filepath.Base(f.FileName)] = p
	}

	return result
}

// insta360StackPhotoIDs returns the distinct photo IDs in owners.
func insta360StackPhotoIDs(owners map[string]entity.Photo) map[uint]bool {
	result := make(map[uint]bool, len(owners))

	for _, p := range owners {
		result[p.ID] = true
	}

	return result
}

// insta360StackPhotoCount returns the number of photo rows in folder, including deleted rows.
func insta360StackPhotoCount(t *testing.T, folder string) (count int) {
	t.Helper()
	require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_path = ?", folder).Count(&count).Error)
	return count
}

// TestIndex_Insta360LateCapture verifies that capture files indexed in a later run are stacked
// with the existing photo, which keeps its name and stays in the archive.
func TestIndex_Insta360LateCapture(t *testing.T) {
	cases := []struct {
		name  string
		first []string
		late  []string
	}{
		{"LateProxy", []string{insta360StackLeft, insta360StackRight}, []string{insta360StackProxy}},
		{"LateLens", []string{insta360StackLeft}, []string{insta360StackRight}},
		{"Reverse", []string{insta360StackRight}, []string{insta360StackLeft}},
	}

	for _, sequences := range []struct {
		name    string
		enabled bool
	}{{"SequencesOff", false}, {"SequencesOn", true}} {
		for _, tc := range cases {
			t.Run(tc.name+"/"+sequences.name, func(t *testing.T) {
				folder := strings.ToLower("insta360" + tc.name + sequences.name)
				cfg := newInsta360StackConfig(t, folder, sequences.enabled)
				dir := filepath.Join(cfg.OriginalsPath(), folder)

				for _, name := range tc.first {
					writeInsta360StackMedia(t, cfg, dir, name)
				}

				indexInsta360StackFolder(cfg, folder, false, true)

				owners := insta360StackOwners(t, folder)
				require.Len(t, owners, len(tc.first))
				require.Len(t, insta360StackPhotoIDs(owners), 1)

				var photo entity.Photo
				for _, p := range owners {
					photo = p
				}

				assert.Equal(t, insta360StackName, photo.PhotoName)
				require.NoError(t, photo.Archive())

				for _, name := range tc.late {
					writeInsta360StackMedia(t, cfg, dir, name)
				}

				// A run that skips archived photos leaves the late file for a later run.
				for _, stage := range []struct {
					name         string
					rescan       bool
					skipArchived bool
					files        int
				}{
					{"SkipArchived", false, true, len(tc.first)},
					{"Incremental", false, false, len(tc.first) + len(tc.late)},
					{"Rescan", true, false, len(tc.first) + len(tc.late)},
				} {
					indexInsta360StackFolder(cfg, folder, stage.rescan, stage.skipArchived)

					owners = insta360StackOwners(t, folder)
					assert.Len(t, owners, stage.files, stage.name)
					assert.Equal(t, map[uint]bool{photo.ID: true}, insta360StackPhotoIDs(owners), stage.name)

					var result entity.Photo
					require.NoError(t, entity.UnscopedDb().First(&result, "id = ?", photo.ID).Error)
					assert.Equal(t, insta360StackName, result.PhotoName, stage.name)
					assert.NotNil(t, result.DeletedAt, stage.name)
					assert.GreaterOrEqual(t, result.PhotoQuality, 0, stage.name)
					assert.Equal(t, 1, insta360StackPhotoCount(t, folder), stage.name)
				}
			})
		}
	}
}

// TestIndexMain_ReplacedPreview verifies that a preview replaced with a forced conversion leaves the file
// cache, so that it is indexed again whatever its modification time.
func TestIndexMain_ReplacedPreview(t *testing.T) {
	folder := "insta360replacedpreview"
	cfg := newInsta360StackConfig(t, folder, false)
	dir := filepath.Join(cfg.OriginalsPath(), folder)
	previewName := folder + "/" + insta360StackLeft + ".jpg"

	writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
	indexInsta360StackFolder(cfg, folder, false, true)
	writeInsta360StackMedia(t, cfg, dir, insta360StackRight)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	require.NoError(t, ind.files.Init())
	require.True(t, ind.files.Exists(previewName, entity.RootSidecar))

	left, err := NewMediaFile(filepath.Join(dir, insta360StackLeft))
	require.NoError(t, err)
	right, err := NewMediaFile(filepath.Join(dir, insta360StackRight))
	require.NoError(t, err)

	related := &RelatedFiles{Main: left, Files: MediaFiles{left, right}}
	result := IndexMain(related, ind, NewIndexOptions(folder, false, true, true, false, true, cfg))
	require.NoError(t, result.Err)

	// IndexMain indexes only the main file, so the replaced preview must no longer be cached.
	var replaced bool

	// Only the replacement made from both lenses has the 2:1 shape of an equirectangular image.
	for _, f := range related.Files[2:] {
		replaced = replaced || f.RootRelName() == previewName && f.InSidecar() && f.Width() == 2*f.Height()
	}

	require.True(t, replaced, "the preview must have been replaced")
	assert.False(t, ind.files.Exists(previewName, entity.RootSidecar))
}

// TestIndex_Insta360StackControls verifies that stacking of other files and sidecar renaming
// follow the actual file names.
func TestIndex_Insta360StackControls(t *testing.T) {
	t.Run("SameBase", func(t *testing.T) {
		folder := "insta360controlsamebase"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, "IMG_1234.jpg")
		indexInsta360StackFolder(cfg, folder, false, true)

		owners := insta360StackOwners(t, folder)
		require.Len(t, owners, 1)
		photo := owners["IMG_1234.jpg"]
		require.NoError(t, photo.Archive())

		writeInsta360StackMedia(t, cfg, dir, "IMG_1234.mp4")
		indexInsta360StackFolder(cfg, folder, false, true)
		assert.Len(t, insta360StackOwners(t, folder), 1, "SkipArchived")

		indexInsta360StackFolder(cfg, folder, false, false)

		owners = insta360StackOwners(t, folder)
		assert.Len(t, owners, 2)
		assert.Equal(t, map[uint]bool{photo.ID: true}, insta360StackPhotoIDs(owners))

		var result entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&result, "id = ?", photo.ID).Error)
		assert.Equal(t, "IMG_1234", result.PhotoName)
		assert.Equal(t, entity.IsStackable, result.PhotoStack)
		assert.NotNil(t, result.DeletedAt)
	})
	t.Run("PartialYaml", func(t *testing.T) {
		folder := "insta360controlpartialyaml"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		// A backup without a UID still applies its values to a new photo.
		writeInsta360StackMedia(t, cfg, dir, "IMG_1234.jpg")
		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_1234.yml"), []byte("Favorite: true\n"), fs.ModeFile))
		indexInsta360StackFolder(cfg, folder, false, true)

		assert.True(t, insta360StackOwners(t, folder)["IMG_1234.jpg"].PhotoFavorite)
	})
	t.Run("LookAlike", func(t *testing.T) {
		folder := "insta360controllookalike"
		cfg := newInsta360StackConfig(t, folder, true)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, "VID_20220625_140410_00_008.mp4")
		indexInsta360StackFolder(cfg, folder, false, true)
		writeInsta360StackMedia(t, cfg, dir, "VID_20220625_140410_10_008.mp4")
		indexInsta360StackFolder(cfg, folder, false, true)

		owners := insta360StackOwners(t, folder)
		require.Len(t, owners, 2)
		assert.Len(t, insta360StackPhotoIDs(owners), 2)
		assert.Equal(t, "VID_20220625_140410_00_008", owners["VID_20220625_140410_00_008.mp4"].PhotoName)
		assert.Equal(t, entity.IsStackable, owners["VID_20220625_140410_00_008.mp4"].PhotoStack)
		assert.Equal(t, "VID_20220625_140410_10_008", owners["VID_20220625_140410_10_008.mp4"].PhotoName)
	})
	t.Run("MovedLens", func(t *testing.T) {
		folder := "insta360controlmoved"
		movedFolder := "insta360controlmovedto"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		movedDir := filepath.Join(cfg.OriginalsPath(), movedFolder)

		// The right lens gets sidecars of its own only while it is indexed without its partner.
		for _, name := range []string{insta360StackRight, insta360StackLeft} {
			writeInsta360StackMedia(t, cfg, dir, name)
			indexInsta360StackFolder(cfg, folder, false, true)
		}

		photoIDs := insta360StackPhotoIDs(insta360StackOwners(t, folder))
		require.Len(t, photoIDs, 1)

		leftSidecars, err := filepath.Glob(filepath.Join(cfg.SidecarPath(), folder, "VID_20220625_140410_00_008.*"))
		require.NoError(t, err)
		rightSidecars, err := filepath.Glob(filepath.Join(cfg.SidecarPath(), folder, "VID_20220625_140410_10_008.*"))
		require.NoError(t, err)
		require.NotEmpty(t, leftSidecars)
		require.NotEmpty(t, rightSidecars)

		require.NoError(t, fs.MkdirAll(movedDir))
		require.NoError(t, fs.Move(filepath.Join(dir, insta360StackRight), filepath.Join(movedDir, insta360StackRight), false))
		indexInsta360StackFolder(cfg, movedFolder, false, true)

		// The moved lens keeps its own sidecars under its own name.
		for _, fileName := range rightSidecars {
			assert.NoFileExists(t, fileName)
			assert.FileExists(t, filepath.Join(cfg.SidecarPath(), movedFolder, filepath.Base(fileName)))
		}

		for _, fileName := range leftSidecars {
			assert.FileExists(t, fileName)
		}

		moved := insta360StackOwners(t, movedFolder)
		assert.Len(t, moved, 1)
		assert.Equal(t, photoIDs, insta360StackPhotoIDs(moved))
	})
	t.Run("MovedLensOfPair", func(t *testing.T) {
		folder := "insta360controlmovedpair"
		movedFolder := "insta360controlmovedpairto"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		movedDir := filepath.Join(cfg.OriginalsPath(), movedFolder)

		for _, name := range []string{insta360StackLeft, insta360StackRight} {
			writeInsta360StackMedia(t, cfg, dir, name)
		}

		indexInsta360StackFolder(cfg, folder, false, true)
		photoIDs := insta360StackPhotoIDs(insta360StackOwners(t, folder))
		require.Len(t, photoIDs, 1)

		leftSidecars, err := filepath.Glob(filepath.Join(cfg.SidecarPath(), folder, "VID_20220625_140410_00_008.*"))
		require.NoError(t, err)
		require.NotEmpty(t, leftSidecars)

		require.NoError(t, fs.MkdirAll(movedDir))
		require.NoError(t, fs.Move(filepath.Join(dir, insta360StackRight), filepath.Join(movedDir, insta360StackRight), false))
		indexInsta360StackFolder(cfg, movedFolder, false, true)

		for _, fileName := range leftSidecars {
			assert.FileExists(t, fileName)
		}

		moved := insta360StackOwners(t, movedFolder)
		assert.Len(t, moved, 1)
		assert.Equal(t, photoIDs, insta360StackPhotoIDs(moved))
	})
}

// TestIndex_Insta360StackOptions verifies stacking of late capture files with other index options
// and photo states, and the recovery of backups written under the stack name.
func TestIndex_Insta360StackOptions(t *testing.T) {
	t.Run("StackDisabled", func(t *testing.T) {
		folder := "insta360optionsstackdisabled"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
		indexInsta360StackFolder(cfg, folder, false, true)
		photoIDs := insta360StackPhotoIDs(insta360StackOwners(t, folder))
		require.Len(t, photoIDs, 1)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackOptions(cfg, NewIndexOptions(folder, false, true, false, false, true, cfg))

		owners := insta360StackOwners(t, folder)
		assert.Len(t, owners, 2)
		assert.Equal(t, photoIDs, insta360StackPhotoIDs(owners))
	})
	t.Run("UnstackedPhoto", func(t *testing.T) {
		folder := "insta360optionsunstacked"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
		indexInsta360StackFolder(cfg, folder, false, true)
		photo := insta360StackOwners(t, folder)[insta360StackLeft]
		require.NoError(t, photo.Update("PhotoStack", entity.IsUnstacked))
		require.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", photo.ID).Error)
		require.Equal(t, entity.IsUnstacked, photo.PhotoStack)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)

		owners := insta360StackOwners(t, folder)
		assert.Len(t, owners, 2)
		assert.Equal(t, map[uint]bool{photo.ID: true}, insta360StackPhotoIDs(owners))
		assert.Equal(t, 1, insta360StackPhotoCount(t, folder))

		// An explicit unstacked flag is kept.
		require.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", photo.ID).Error)
		assert.Equal(t, entity.IsUnstacked, photo.PhotoStack)
	})
	t.Run("YamlRestore", func(t *testing.T) {
		folder := "insta360optionsyamlrestore"
		cfg := newInsta360StackConfig(t, folder, false)
		cfg.Options().SidecarYaml = true
		cfg.Options().DisableBackups = false
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)

		photo := insta360StackOwners(t, folder)[insta360StackRight]
		require.True(t, photo.HasUID())
		require.Equal(t, insta360StackName, photo.PhotoName)
		require.FileExists(t, filepath.Join(cfg.SidecarPath(), folder, insta360StackName+".yml"))

		require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.File{}).Error)
		require.NoError(t, entity.UnscopedDb().Where("id = ?", photo.ID).Delete(&entity.Photo{}).Error)

		indexInsta360StackFolder(cfg, folder, false, true)

		assert.Equal(t, photo.PhotoUID, insta360StackOwners(t, folder)[insta360StackRight].PhotoUID)
	})
	t.Run("SplitYamlRestore", func(t *testing.T) {
		folder := "insta360optionssplityaml"
		cfg := newInsta360StackConfig(t, folder, false)
		cfg.Options().SidecarYaml = true
		cfg.Options().DisableBackups = false
		// The split files restore different photos, so the result depends on the order they are indexed in.
		cfg.Options().IndexWorkers = "1"
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		sidecarDir := filepath.Join(cfg.SidecarPath(), folder)

		// Create a split capture with one backup per photo, named after its own file.
		writeInsta360StackMedia(t, cfg, dir, insta360StackProxy)
		indexInsta360StackFolder(cfg, folder, false, true)
		proxy := insta360StackOwners(t, folder)[insta360StackProxy]
		require.NoError(t, proxy.Update("PhotoName", "LRV_20220625_140410_11_008"))
		require.NoError(t, entity.UnscopedDb().First(&proxy, "id = ?", proxy.ID).Error)
		require.NoError(t, proxy.SaveSidecarYaml(cfg.OriginalsPath(), cfg.SidecarPath()))
		require.NoError(t, os.Remove(filepath.Join(sidecarDir, insta360StackName+".yml")))
		require.FileExists(t, filepath.Join(sidecarDir, "LRV_20220625_140410_11_008.yml"))

		writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
		indexInsta360StackFolder(cfg, folder, false, true)
		owners := insta360StackOwners(t, folder)
		left := owners[insta360StackLeft]
		require.Len(t, insta360StackPhotoIDs(owners), 2)
		require.FileExists(t, filepath.Join(sidecarDir, insta360StackName+".yml"))

		for id := range insta360StackPhotoIDs(owners) {
			require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", id).Delete(&entity.File{}).Error)
			require.NoError(t, entity.UnscopedDb().Where("id = ?", id).Delete(&entity.Photo{}).Error)
		}

		indexInsta360StackFolder(cfg, folder, false, true)

		owners = insta360StackOwners(t, folder)
		assert.Equal(t, proxy.PhotoUID, owners[insta360StackProxy].PhotoUID)
		assert.Equal(t, left.PhotoUID, owners[insta360StackLeft].PhotoUID)
		assert.Equal(t, folder, owners[insta360StackProxy].PhotoPath)
		assert.Equal(t, "LRV_20220625_140410_11_008", owners[insta360StackProxy].PhotoName)
		assert.Equal(t, insta360StackName, owners[insta360StackLeft].PhotoName)
		assert.FileExists(t, filepath.Join(sidecarDir, "LRV_20220625_140410_11_008.yml"))
		assert.NoFileExists(t, filepath.Join(cfg.SidecarPath(), "LRV_20220625_140410_11_008.yml"))
	})
	t.Run("YamlWithoutUID", func(t *testing.T) {
		folder := "insta360optionsyamlwithoutuid"
		cfg := newInsta360StackConfig(t, folder, false)
		cfg.Options().SidecarYaml = true
		cfg.Options().DisableBackups = false
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)
		photo := insta360StackOwners(t, folder)[insta360StackRight]
		require.True(t, photo.HasUID())

		// A backup under the file's own name without a UID does not hide the photo's backup.
		require.NoError(t, os.WriteFile(filepath.Join(dir, "VID_20220625_140410_10_008.yml"), []byte("Favorite: true\n"), fs.ModeFile))
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.File{}).Error)
		require.NoError(t, entity.UnscopedDb().Where("id = ?", photo.ID).Delete(&entity.Photo{}).Error)

		indexInsta360StackFolder(cfg, folder, false, true)

		restored := insta360StackOwners(t, folder)[insta360StackRight]
		assert.Equal(t, photo.PhotoUID, restored.PhotoUID)
		assert.False(t, restored.PhotoFavorite)
		assert.Equal(t, insta360StackName, restored.PhotoName)
		assert.Equal(t, folder, restored.PhotoPath)
	})
	t.Run("ExistingName", func(t *testing.T) {
		folder := "insta360optionsexistingname"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)

		// A photo stored under the file's own base name keeps it on a forced rescan.
		photo := insta360StackOwners(t, folder)[insta360StackRight]
		require.NoError(t, photo.Update("PhotoName", "VID_20220625_140410_10_008"))

		indexInsta360StackFolder(cfg, folder, true, false)

		var result entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&result, "id = ?", photo.ID).Error)
		assert.Equal(t, "VID_20220625_140410_10_008", result.PhotoName)
	})
	t.Run("NameLock", func(t *testing.T) {
		folder := "insta360optionsnamelock"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
		indexInsta360StackFolder(cfg, folder, false, true)
		photoIDs := insta360StackPhotoIDs(insta360StackOwners(t, folder))
		require.Len(t, photoIDs, 1)

		// A new file waits for the lock on its stack name before looking up the photo.
		writeInsta360StackMedia(t, cfg, dir, insta360StackProxy)
		unlock := lockStackName(folder, insta360StackName)
		done := make(chan struct{})

		go func() {
			defer close(done)
			indexInsta360StackFolder(cfg, folder, false, true)
		}()

		select {
		case <-done:
			unlock()
			t.Fatal("indexing finished while the stack name was locked")
		case <-time.After(3 * time.Second):
		}

		assert.Len(t, insta360StackOwners(t, folder), 1)
		unlock()
		<-done

		owners := insta360StackOwners(t, folder)
		assert.Len(t, owners, 2)
		assert.Equal(t, photoIDs, insta360StackPhotoIDs(owners))
	})
}

// TestIndex_Insta360LensCodedPhotos verifies that photos with lens codes are indexed as separate photos,
// since Insta360 cameras never split a photo by lens.
func TestIndex_Insta360LensCodedPhotos(t *testing.T) {
	const (
		left  = "IMG_20220625_140410_00_008.insp"
		right = "IMG_20220625_140410_10_008.insp"
	)

	cases := []struct {
		name  string
		first []string
		late  []string
	}{
		{"SameRun", []string{left, right}, nil},
		{"LateRight", []string{left}, []string{right}},
		{"Reverse", []string{right}, []string{left}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			folder := strings.ToLower("insta360photo" + tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			for _, name := range tc.first {
				writeInsta360StackMedia(t, cfg, dir, name)
			}

			indexInsta360StackFolder(cfg, folder, false, true)

			for _, name := range tc.late {
				writeInsta360StackMedia(t, cfg, dir, name)
			}

			if len(tc.late) > 0 {
				indexInsta360StackFolder(cfg, folder, false, true)
			}

			owners := insta360StackOwners(t, folder)
			assert.Len(t, owners, 2)
			assert.Len(t, insta360StackPhotoIDs(owners), 2)
			assert.Equal(t, "IMG_20220625_140410_00_008", owners[left].PhotoName)
			assert.Equal(t, "IMG_20220625_140410_10_008", owners[right].PhotoName)
		})
	}
}

// TestIndex_Insta360StandaloneLensPhoto verifies that a single-lens photo named with the _10 lens
// code is indexed as an ordinary photo.
func TestIndex_Insta360StandaloneLensPhoto(t *testing.T) {
	folder := "insta360standalonelens"
	cfg := newInsta360StackConfig(t, folder, false)
	dir := filepath.Join(cfg.OriginalsPath(), folder)
	name := "IMG_20231015_101112_10_124.insp"

	require.NoError(t, fs.MkdirAll(dir))
	require.NoError(t, fs.Copy("testdata/insta360.insp", filepath.Join(dir, name), false))
	indexInsta360StackFolder(cfg, folder, false, true)

	owners := insta360StackOwners(t, folder)
	require.Len(t, owners, 1)
	photo := owners[name]
	assert.Equal(t, "IMG_20231015_101112_10_124", photo.PhotoName)
	assert.Equal(t, entity.IsStackable, photo.PhotoStack)
	assert.GreaterOrEqual(t, photo.PhotoQuality, 0)

	var file entity.File
	require.NoError(t, entity.UnscopedDb().First(&file, "file_name = ?", folder+"/"+name).Error)
	assert.False(t, file.KeepStacked())
	assert.Equal(t, "", file.StackGroup())
}

// TestIndex_Insta360ImportedName verifies that a capture file renamed on import keeps its original
// name, so it is still identified as a file that must stay stacked.
func TestIndex_Insta360ImportedName(t *testing.T) {
	folder := "insta360importedname"
	cfg := newInsta360StackConfig(t, folder, false)
	dir := filepath.Join(cfg.OriginalsPath(), folder)

	writeInsta360StackMedia(t, cfg, dir, "20260925_135937_07784009.insv")
	mediaFile, err := NewMediaFile(filepath.Join(dir, "20260925_135937_07784009.insv"))
	require.NoError(t, err)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	result := ind.UserMediaFile(mediaFile, NewIndexOptions(folder, false, true, true, false, true, cfg), insta360StackRight, "", entity.OwnerUnknown)
	require.True(t, result.Success(), result.Err)

	var file entity.File
	require.NoError(t, entity.UnscopedDb().First(&file, "file_name = ?", folder+"/20260925_135937_07784009.insv").Error)
	assert.Equal(t, insta360StackRight, file.OriginalName)
	assert.True(t, file.KeepStacked())
	assert.Equal(t, insta360StackName, file.StackGroup())
	assert.Equal(t, entity.IsStackable, insta360StackOwners(t, folder)["20260925_135937_07784009.insv"].PhotoStack)
}

// splitInsta360Capture moves a capture file and its previews to a new photo, and names both photos
// after their own files, which is how split captures are stored.
func splitInsta360Capture(t *testing.T, existing entity.Photo, folder, first, late string) entity.Photo {
	t.Helper()

	require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", existing.ID).
		UpdateColumn("photo_name", fs.StripKnownExt(first)).Error)

	photo := entity.NewPhoto(true)
	photo.PhotoPath = folder
	photo.PhotoName = fs.StripKnownExt(late)
	photo.PhotoType = entity.MediaVideo
	photo.PhotoQuality = 3
	require.NoError(t, photo.Create())

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().
		Where("photo_id = ? AND file_name LIKE ?", existing.ID, folder+"/"+late+"%").
		Find(&files).Error)
	require.NotEmpty(t, files)

	for _, f := range files {
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", f.ID).UpdateColumns(entity.Values{
			"photo_id":     photo.ID,
			"photo_uid":    photo.PhotoUID,
			"file_primary": f.FileType == fs.ImageJpeg.String(),
		}).Error)
	}

	return photo
}

// TestIndex_Insta360ReconcileSplit verifies that a forced rescan merges a split capture into the
// existing photo, whose archive state is kept unless it was removed automatically.
func TestIndex_Insta360ReconcileSplit(t *testing.T) {
	const (
		active = iota
		archived
		removed
	)

	cases := []struct {
		name     string
		first    string
		late     string
		existing int
		extra    int
		archived bool
	}{
		{"Archived", insta360StackLeft, insta360StackRight, archived, active, true},
		{"ArchivedReverse", insta360StackRight, insta360StackLeft, archived, active, true},
		{"MemberArchived", insta360StackLeft, insta360StackRight, active, archived, false},
		{"Removed", insta360StackLeft, insta360StackRight, removed, removed, false},
		{"RemovedReverse", insta360StackRight, insta360StackLeft, removed, removed, false},
		{"RemovedMemberArchived", insta360StackLeft, insta360StackRight, removed, archived, true},
	}

	setState := func(t *testing.T, p entity.Photo, state int) {
		switch state {
		case archived:
			require.NoError(t, p.Archive())
		case removed:
			_, err := p.Delete(false)
			require.NoError(t, err)
		}
	}

	for _, tc := range cases {
		for _, skipArchived := range []bool{false, true} {
			stage := "Archived"
			if skipArchived {
				stage = "SkipArchived"
			}

			t.Run(tc.name+"/"+stage, func(t *testing.T) {
				folder := strings.ToLower("insta360split" + tc.name + stage)
				cfg := newInsta360StackConfig(t, folder, false)
				dir := filepath.Join(cfg.OriginalsPath(), folder)

				writeInsta360StackMedia(t, cfg, dir, tc.first)
				indexInsta360StackFolder(cfg, folder, false, true)
				writeInsta360StackMedia(t, cfg, dir, tc.late)
				indexInsta360StackFolder(cfg, folder, false, true)

				owners := insta360StackOwners(t, folder)
				require.Len(t, insta360StackPhotoIDs(owners), 1)
				existing := owners[tc.first]
				extra := splitInsta360Capture(t, existing, folder, tc.first, tc.late)
				require.Len(t, insta360StackPhotoIDs(insta360StackOwners(t, folder)), 2)

				setState(t, existing, tc.existing)
				setState(t, extra, tc.extra)

				indexInsta360StackFolder(cfg, folder, true, skipArchived)

				var photoIDs []uint
				require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).
					Where("file_name LIKE ?", folder+"/%").Pluck("DISTINCT photo_id", &photoIDs).Error)
				assert.Equal(t, []uint{existing.ID}, photoIDs)

				owners = insta360StackOwners(t, folder)
				assert.Len(t, owners, 2)

				result := owners[tc.first]
				assert.Equal(t, tc.archived, result.DeletedAt != nil)
				assert.GreaterOrEqual(t, result.PhotoQuality, 0)
			})
		}
	}
}

// insta360StackPreviews returns the preview file rows in folder, keyed by file name.
func insta360StackPreviews(t *testing.T, folder string) map[string]entity.File {
	t.Helper()

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().
		Where("file_name LIKE ? AND file_type = ?", folder+"/%", fs.ImageJpeg.String()).
		Find(&files).Error)

	result := make(map[string]entity.File, len(files))

	for _, f := range files {
		result[filepath.Base(f.FileName)] = f
	}

	return result
}

// assertInsta360SinglePrimary checks that all files in folder belong to one photo with one primary file.
func assertInsta360SinglePrimary(t *testing.T, folder string) {
	t.Helper()

	var photoIDs []uint
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).
		Where("file_name LIKE ?", folder+"/%").Pluck("DISTINCT photo_id", &photoIDs).Error)
	require.Len(t, photoIDs, 1)

	var primaries int
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).
		Where("photo_id = ? AND file_primary = 1", photoIDs[0]).Count(&primaries).Error)
	assert.Equal(t, 1, primaries)
}

// TestIndex_Insta360Cover verifies that the combined preview becomes the cover of a capture whose
// lens files were indexed in separate runs, and that no preview is made from the right lens then.
func TestIndex_Insta360Cover(t *testing.T) {
	const (
		leftPreview  = insta360StackLeft + ".jpg"
		rightPreview = insta360StackRight + ".jpg"
	)

	cases := []struct {
		name         string
		first        string
		late         string
		archive      bool
		rightPreview bool
	}{
		{"LateLens", insta360StackLeft, insta360StackRight, false, false},
		{"LateLensArchived", insta360StackLeft, insta360StackRight, true, false},
		{"Reverse", insta360StackRight, insta360StackLeft, false, true},
		{"ReverseArchived", insta360StackRight, insta360StackLeft, true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			folder := strings.ToLower("insta360cover" + tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			writeInsta360StackMedia(t, cfg, dir, tc.first)
			indexInsta360StackFolder(cfg, folder, false, true)

			if tc.archive {
				photo := insta360StackOwners(t, folder)[tc.first]
				require.NoError(t, photo.Archive())
			}

			// A preview replaced in the same run is indexed again even with the same whole-second time. A run
			// that skips the archived photo leaves its row to a later run, which compares the recorded time.
			sameSecond := !tc.archive && !tc.rightPreview
			var next time.Time

			if sameSecond {
				next = sameSecondInsta360Previews(t, cfg, folder)
			} else {
				backdateInsta360Previews(t, filepath.Join(cfg.SidecarPath(), folder))
			}

			writeInsta360StackMedia(t, cfg, dir, tc.late)
			indexInsta360StackFolder(cfg, folder, false, true)

			// A slow run writes the preview in a later second, which the check below then cannot tell apart.
			if stat, statErr := os.Stat(filepath.Join(cfg.SidecarPath(), folder, leftPreview)); sameSecond && statErr == nil &&
				stat.ModTime().Unix() != next.Unix() {
				t.Logf("preview was replaced after %s, so the same-second case is not covered", next.Format(time.RFC3339))
			}

			if tc.archive {
				indexInsta360StackFolder(cfg, folder, false, false)
			}

			previews := insta360StackPreviews(t, folder)
			require.Contains(t, previews, leftPreview)
			assert.True(t, previews[leftPreview].FilePrimary)
			assert.Equal(t, "equirectangular", previews[leftPreview].FileProjection)

			if tc.rightPreview {
				require.Contains(t, previews, rightPreview)
				assert.False(t, previews[rightPreview].FilePrimary)
				assert.Equal(t, "", previews[rightPreview].FileProjection)
			} else {
				assert.NotContains(t, previews, rightPreview)
				assert.NoFileExists(t, filepath.Join(cfg.SidecarPath(), folder, rightPreview))
			}

			assertInsta360SinglePrimary(t, folder)

			// A forced rescan keeps the cover.
			indexInsta360StackFolder(cfg, folder, true, false)
			previews = insta360StackPreviews(t, folder)
			assert.True(t, previews[leftPreview].FilePrimary)
			assert.Equal(t, "equirectangular", previews[leftPreview].FileProjection)
		})
	}
	t.Run("LabeledRightPreview", func(t *testing.T) {
		folder := "insta360coverlabeled"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_name = ?", folder+"/"+rightPreview).
			UpdateColumn("file_projection", "equirectangular").Error)

		writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
		indexInsta360StackFolder(cfg, folder, false, true)
		indexInsta360StackFolder(cfg, folder, true, false)

		previews := insta360StackPreviews(t, folder)
		assert.Equal(t, "", previews[rightPreview].FileProjection)
		assert.False(t, previews[rightPreview].FilePrimary)
		assert.Equal(t, "equirectangular", previews[leftPreview].FileProjection)
	})
	for _, ffmpegCase := range []struct {
		name    string
		disable func(t *testing.T, cfg *config.Config)
	}{
		{"FFmpegDisabled", func(t *testing.T, cfg *config.Config) { cfg.Options().DisableFFmpeg = true }},
		{"FFmpegExcluded", func(t *testing.T, cfg *config.Config) {
			exclude := ffmpeg.Exclude()
			ffmpeg.SetExclude(video.NewFormats(fs.VideoInsv.String()))
			t.Cleanup(func() { ffmpeg.SetExclude(exclude) })
		}},
	} {
		t.Run(ffmpegCase.name, func(t *testing.T) {
			folder := strings.ToLower("insta360cover" + ffmpegCase.name)
			cfg := newInsta360StackConfig(t, folder, false)
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
			indexInsta360StackFolder(cfg, folder, false, true)
			ffmpegCase.disable(t, cfg)
			writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
			indexInsta360StackFolder(cfg, folder, false, true)

			// The existing preview is kept, and the right lens is indexed.
			previews := insta360StackPreviews(t, folder)
			require.Contains(t, previews, leftPreview)
			assert.True(t, previews[leftPreview].FilePrimary)
			assert.FileExists(t, filepath.Join(cfg.SidecarPath(), folder, leftPreview))
			assert.Len(t, insta360StackOwners(t, folder), 2)
			assertInsta360SinglePrimary(t, folder)
		})
	}
	t.Run("OtherPrimary", func(t *testing.T) {
		folder := "insta360coverotherprimary"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)
		photo := insta360StackOwners(t, folder)[insta360StackRight]

		// A picture the user stacked with the capture and chose as its cover.
		writeInsta360StackMedia(t, cfg, dir, "cover.jpg")
		indexInsta360StackFolder(cfg, folder, false, true)
		var cover entity.File
		require.NoError(t, entity.UnscopedDb().First(&cover, "file_name = ?", folder+"/cover.jpg").Error)
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", cover.ID).
			UpdateColumns(entity.Values{"photo_id": photo.ID, "photo_uid": photo.PhotoUID}).Error)
		require.NoError(t, photo.SetPrimary(cover.FileUID))

		writeInsta360StackMedia(t, cfg, dir, insta360StackLeft)
		indexInsta360StackFolder(cfg, folder, false, true)
		indexInsta360StackFolder(cfg, folder, true, false)

		require.NoError(t, entity.UnscopedDb().First(&cover, "id = ?", cover.ID).Error)
		assert.True(t, cover.FilePrimary)
		assert.False(t, insta360StackPreviews(t, folder)[leftPreview].FilePrimary)
		assertInsta360SinglePrimary(t, folder)
	})
	t.Run("SingleLens", func(t *testing.T) {
		folder := "insta360coversinglelens"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, insta360StackRight)
		indexInsta360StackFolder(cfg, folder, false, true)
		indexInsta360StackFolder(cfg, folder, true, false)

		previews := insta360StackPreviews(t, folder)
		require.Len(t, previews, 1)
		assert.True(t, previews[rightPreview].FilePrimary)
		assert.Equal(t, "", previews[rightPreview].FileProjection)
	})
}

// TestIndex_Insta360Proxy verifies that the LRV proxy of a video that stores both lenses in one file
// is only indexed with that video, in any order, is not converted, and must stay stacked.
func TestIndex_Insta360Proxy(t *testing.T) {
	const (
		left  = "VID_20240415_213145_00_035.insv"
		proxy = "LRV_20240415_213145_01_035.lrv"
		name  = "VID_20240415_213145_00_035"
	)

	cases := []struct {
		name  string
		first []string
		late  []string
	}{
		{"SameRun", []string{left, proxy}, nil},
		{"LateProxy", []string{left}, []string{proxy}},
		{"ProxyFirst", []string{proxy}, []string{left}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			folder := strings.ToLower("insta360proxy" + tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			for _, fileName := range tc.first {
				writeInsta360StackMedia(t, cfg, dir, fileName)
			}

			indexInsta360StackFolder(cfg, folder, false, true)

			if len(tc.late) > 0 {
				// A proxy without its video is not indexed.
				indexed := insta360StackOwners(t, folder)
				assert.Equal(t, tc.first[0] == left, len(indexed) == 1)
				assert.NotContains(t, indexed, proxy)

				for _, fileName := range tc.late {
					writeInsta360StackMedia(t, cfg, dir, fileName)
				}

				indexInsta360StackFolder(cfg, folder, false, true)
			}

			owners := insta360StackOwners(t, folder)
			require.Len(t, owners, 2)
			assert.Len(t, insta360StackPhotoIDs(owners), 1)
			assert.Equal(t, name, owners[left].PhotoName)

			var file entity.File
			require.NoError(t, entity.UnscopedDb().First(&file, "file_name = ?", folder+"/"+proxy).Error)
			assert.Equal(t, fs.VideoLrv.String(), file.FileType)
			assert.False(t, file.FilePrimary)
			assert.True(t, file.KeepStacked())
			assert.Equal(t, name, file.StackGroup())

			assert.NoFileExists(t, filepath.Join(cfg.SidecarPath(), folder, proxy+".jpg"))
			assert.True(t, insta360StackPreviews(t, folder)[left+".jpg"].FilePrimary)
			assertInsta360SinglePrimary(t, folder)
		})
	}
	t.Run("IgnoredVideo", func(t *testing.T) {
		folder := "insta360proxyignored"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, left)
		writeInsta360StackMedia(t, cfg, dir, proxy)
		require.NoError(t, os.WriteFile(filepath.Join(dir, fs.PPIgnoreFilename), []byte(left+"\n"), fs.ModeFile))
		indexInsta360StackFolder(cfg, folder, false, true)

		assert.Empty(t, insta360StackOwners(t, folder))
	})
	t.Run("OtherProxies", func(t *testing.T) {
		folder := "insta360proxyother"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		// Proxies of other cameras, and an Insta360 proxy without its video, are not indexed.
		for _, fileName := range []string{"GX010123.MP4", "GL010123.LRV", "GOPR0124.MP4", "GOPR0124.LRV", "LRV_20240415_213145_01_036.lrv"} {
			writeInsta360StackMedia(t, cfg, dir, fileName)
		}

		indexInsta360StackFolder(cfg, folder, false, true)
		indexInsta360StackFolder(cfg, folder, true, false)

		owners := insta360StackOwners(t, folder)
		assert.Len(t, owners, 2)
		assert.Contains(t, owners, "GX010123.MP4")
		assert.Contains(t, owners, "GOPR0124.MP4")

		for _, fileName := range []string{"GL010123.LRV", "GOPR0124.LRV", "LRV_20240415_213145_01_036.lrv"} {
			matches, err := filepath.Glob(filepath.Join(cfg.SidecarPath(), folder, fileName+".*"))
			require.NoError(t, err)
			assert.Empty(t, matches, fileName)
		}
	})
}

// TestMediaFile_RelatedFiles_Insta360Proxy verifies that LRV proxies are grouped by partner only in
// originals, while files to be imported are grouped by name as before.
func TestMediaFile_RelatedFiles_Insta360Proxy(t *testing.T) {
	const (
		left  = "VID_20240415_213145_00_035.insv"
		proxy = "LRV_20240415_213145_01_035.lrv"
	)

	cfg := newInsta360StackConfig(t, "insta360proxyrelated", false)

	names := func(related RelatedFiles) (result []string) {
		for _, f := range related.Files {
			result = append(result, f.BaseName())
		}
		return result
	}

	for _, root := range []struct {
		name       string
		dir        string
		proxyInSet bool
		goProInSet bool
	}{
		{"Originals", filepath.Join(cfg.OriginalsPath(), "insta360proxyrelated"), true, false},
		{"Import", filepath.Join(cfg.ImportPath(), "insta360proxyrelated"), false, true},
	} {
		t.Run(root.name, func(t *testing.T) {
			for _, name := range []string{left, proxy, "GOPR0124.MP4", "GOPR0124.LRV"} {
				writeInsta360StackMedia(t, cfg, root.dir, name)
			}

			main, err := NewMediaFile(filepath.Join(root.dir, left))
			require.NoError(t, err)
			related, err := main.RelatedFiles(false)
			require.NoError(t, err)
			assert.Equal(t, root.proxyInSet, slices.Contains(names(related), proxy))
			assert.Equal(t, left, related.Main.BaseName())

			goPro, err := NewMediaFile(filepath.Join(root.dir, "GOPR0124.MP4"))
			require.NoError(t, err)
			related, err = goPro.RelatedFiles(false)
			require.NoError(t, err)
			assert.Equal(t, root.goProInSet, slices.Contains(names(related), "GOPR0124.LRV"))
			assert.Equal(t, "GOPR0124.MP4", related.Main.BaseName())
		})
	}
}

// writeInsta360Streams writes a one-second video with one stream per size, encoded with codec.
func writeInsta360Streams(t *testing.T, cfg *config.Config, fileName, codec string, sizes ...string) {
	t.Helper()
	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))

	args := []string{"-y", "-loglevel", "error"}
	for i, size := range sizes {
		args = append(args, "-f", "lavfi", "-i", fmt.Sprintf("testsrc=size=%s:rate=10,hue=h=%d", size, i*90))
	}

	for i := range sizes {
		args = append(args, "-map", fmt.Sprintf("%d:v", i))
	}

	args = append(args, "-t", "1", "-c:v", codec, "-pix_fmt", "yuv420p", "-metadata", "title="+filepath.Base(fileName), "-f", "mp4", fileName)

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), args...).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))
}

// newInsta360LogHook captures log entries until the test ends, restoring the previous hooks afterwards.
func newInsta360LogHook(t *testing.T) *test.Hook {
	t.Helper()
	logger := logrus.StandardLogger()
	oldHooks := make(logrus.LevelHooks, len(logger.Hooks))
	for level, hooks := range logger.Hooks {
		oldHooks[level] = append([]logrus.Hook(nil), hooks...)
	}
	t.Cleanup(func() { logger.ReplaceHooks(oldHooks) })
	return test.NewGlobal()
}

// insta360DewarpWarnings returns the warnings about 360° originals that could not be dewarped.
func insta360DewarpWarnings(hook *test.Hook) (result []string) {
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "could not be dewarped") {
			result = append(result, entry.Message)
		}
	}
	return result
}

// TestIndex_Insta360DualStream verifies that an .insv with one stream per lens gets an equirectangular
// preview and video made from both streams, while other videos are processed as before.
func TestIndex_Insta360DualStream(t *testing.T) {
	const name = "VID_20240415_213145_00_035.insv"

	t.Run("TwoStreams", func(t *testing.T) {
		folder := "insta360dualstream"
		cfg := newInsta360StackConfig(t, folder, false)
		writeInsta360Streams(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, name), "libx264", "320x320", "320x320")
		hook := newInsta360LogHook(t)
		indexInsta360StackFolder(cfg, folder, false, true)
		assert.Empty(t, insta360DewarpWarnings(hook))

		previews := insta360StackPreviews(t, folder)
		require.Contains(t, previews, name+".jpg")
		assert.True(t, previews[name+".jpg"].FilePrimary)
		assert.Equal(t, "equirectangular", previews[name+".jpg"].FileProjection)
		assert.Equal(t, 640, previews[name+".jpg"].FileWidth)
		assert.Equal(t, 320, previews[name+".jpg"].FileHeight)

		var avc entity.File
		require.NoError(t, entity.UnscopedDb().First(&avc, "file_name = ?", folder+"/"+name+".avc").Error)
		assert.Equal(t, "equirectangular", avc.FileProjection)
		assert.Equal(t, 2, avc.FileWidth/avc.FileHeight)
	})
	t.Run("OrdinaryTwoTrackVideo", func(t *testing.T) {
		folder := "insta360dualstreammp4"
		cfg := newInsta360StackConfig(t, folder, false)
		writeInsta360Streams(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, "clip.mp4"), "libx264", "320x320", "320x320")
		indexInsta360StackFolder(cfg, folder, false, true)

		previews := insta360StackPreviews(t, folder)
		require.Contains(t, previews, "clip.mp4.jpg")
		assert.Equal(t, "", previews["clip.mp4.jpg"].FileProjection)
		assert.Equal(t, 320, previews["clip.mp4.jpg"].FileWidth)
	})
	t.Run("SingleLensHevc", func(t *testing.T) {
		folder := "insta360dualstreamsingle"
		cfg := newInsta360StackConfig(t, folder, false)
		writeInsta360Streams(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, name), "libx265", "320x320")
		indexInsta360StackFolder(cfg, folder, false, true)

		// One square lens is not dewarped as both lenses, even if its dimensions come from the track header.
		previews := insta360StackPreviews(t, folder)
		require.Contains(t, previews, name+".jpg")
		assert.Equal(t, "", previews[name+".jpg"].FileProjection)
		assert.Equal(t, 320, previews[name+".jpg"].FileWidth)
	})
	t.Run("ExistingPreview", func(t *testing.T) {
		folder := "insta360dualstreamexisting"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		writeInsta360Streams(t, cfg, filepath.Join(dir, name), "libx264", "320x320", "320x320")
		require.NoError(t, fs.Copy("testdata/flash.jpg", filepath.Join(dir, name+".jpg"), false))

		// A preview the user added is kept, and is not reported as a failed dewarp.
		hook := newInsta360LogHook(t)
		indexInsta360StackFolder(cfg, folder, true, false)
		assert.Empty(t, insta360DewarpWarnings(hook))
		assert.FileExists(t, filepath.Join(dir, name+".jpg"))
	})
	t.Run("DewarpFailed", func(t *testing.T) {
		folder := "insta360dualstreamfailed"
		cfg := newInsta360StackConfig(t, folder, false)
		writeInsta360Streams(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, name), "libx264", "320x320", "320x320")

		// A wrapper that fails every dewarp, so the preview falls back to a still frame.
		wrapper := filepath.Join(t.TempDir(), "ffmpeg")
		require.NoError(t, os.WriteFile(wrapper, []byte("#!/bin/sh\ncase \"$*\" in *v360*) exit 1;; esac\nexec "+cfg.FFmpegBin()+" \"$@\"\n"), 0o700))
		cfg.Options().FFmpegBin = wrapper

		hook := newInsta360LogHook(t)
		indexInsta360StackFolder(cfg, folder, false, true)

		previews := insta360StackPreviews(t, folder)
		require.Contains(t, previews, name+".jpg")
		assert.Equal(t, "", previews[name+".jpg"].FileProjection)
		assert.Len(t, insta360DewarpWarnings(hook), 1)
	})
}

// writeInsta360Photo writes a JPEG of the specified size under an .insp name, with bytes that differ per name.
func writeInsta360Photo(t *testing.T, cfg *config.Config, fileName, size string) {
	t.Helper()
	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), "-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size="+size,
		"-frames:v", "1", "-f", "image2", "-c:v", "mjpeg", "-metadata", "comment="+filepath.Base(fileName), fileName).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))
}

// TestIndex_Insta360SingleLensPhoto verifies that an .insp with a single lens is neither dewarped nor
// labeled as dual-fisheye, while an .insp with both lenses side by side, which is exactly 2:1, still is.
func TestIndex_Insta360SingleLensPhoto(t *testing.T) {
	const name = "IMG_20201026_154628_00_070.insp"

	original := func(t *testing.T, folder string) (result entity.File) {
		t.Helper()
		require.NoError(t, entity.UnscopedDb().First(&result, "file_root = ? AND file_name = ?", entity.RootOriginals, folder+"/"+name).Error)
		return result
	}

	// Single-lens photo modes use these shapes; 2.1:1 is close to 2:1 but still not a dual-lens frame.
	for _, tc := range []struct {
		name, size string
		width      int
		height     int
	}{
		{"Wide", "320x180", 320, 180},
		{"Tall", "180x320", 180, 320},
		{"Square", "320x320", 320, 320},
		{"Panorama", "480x160", 480, 160},
		{"NearlyDual", "336x160", 336, 160},
	} {
		t.Run(tc.name, func(t *testing.T) {
			folder := strings.ToLower("insta360singlelens" + tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			writeInsta360Photo(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, name), tc.size)

			hook := newInsta360LogHook(t)
			indexInsta360StackFolder(cfg, folder, false, true)
			assert.Empty(t, insta360DewarpWarnings(hook))
			assert.Equal(t, "", original(t, folder).FileProjection)

			previews := insta360StackPreviews(t, folder)
			require.Contains(t, previews, name+".jpg")
			assert.Equal(t, "", previews[name+".jpg"].FileProjection)
			assert.Equal(t, tc.width, previews[name+".jpg"].FileWidth)
			assert.Equal(t, tc.height, previews[name+".jpg"].FileHeight)

			// A label stored by an earlier index run is cleared.
			require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", original(t, folder).ID).
				UpdateColumn("file_projection", "dual-fisheye").Error)
			indexInsta360StackFolder(cfg, folder, true, false)
			assert.Equal(t, "", original(t, folder).FileProjection)
			assert.Equal(t, "", insta360StackPreviews(t, folder)[name+".jpg"].FileProjection)
		})
	}
	t.Run("DualLens", func(t *testing.T) {
		folder := "insta360duallensphoto"
		cfg := newInsta360StackConfig(t, folder, false)
		writeInsta360Photo(t, cfg, filepath.Join(cfg.OriginalsPath(), folder, name), "320x160")
		indexInsta360StackFolder(cfg, folder, false, true)

		assert.Equal(t, "dual-fisheye", original(t, folder).FileProjection)
		assert.Equal(t, "equirectangular", insta360StackPreviews(t, folder)[name+".jpg"].FileProjection)
	})
	t.Run("SideBySideFixture", func(t *testing.T) {
		folder := "insta360sidebysidephoto"
		cfg := newInsta360StackConfig(t, folder, false)
		dir := filepath.Join(cfg.OriginalsPath(), folder)
		require.NoError(t, fs.MkdirAll(dir))
		require.NoError(t, fs.Copy("testdata/insta360.insp", filepath.Join(dir, name), false))
		indexInsta360StackFolder(cfg, folder, false, true)

		assert.Equal(t, "dual-fisheye", original(t, folder).FileProjection)
		assert.Equal(t, "equirectangular", insta360StackPreviews(t, folder)[name+".jpg"].FileProjection)
	})
}

// TestImport_Insta360Capture verifies that a capture whose files are renamed on import gets the combined
// preview, and that a forced rescan replaces a single-lens preview of an earlier import.
func TestImport_Insta360Capture(t *testing.T) {
	folder := "insta360import"
	cfg := newInsta360StackConfig(t, folder, false)
	importDir := filepath.Join(cfg.ImportPath(), folder)

	for _, name := range []string{insta360StackLeft, insta360StackRight, insta360StackProxy} {
		writeInsta360StackMedia(t, cfg, importDir, name)
	}

	convert := NewConvert(cfg)
	NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert).Start(ImportOptionsMove(importDir, folder))

	stored := func(originalName string) (result entity.File) {
		t.Helper()
		require.NoError(t, entity.UnscopedDb().First(&result, "file_root = ? AND original_name = ?", entity.RootOriginals, originalName).Error)
		return result
	}

	left, right, proxy := stored(insta360StackLeft), stored(insta360StackRight), stored(insta360StackProxy)
	require.NotEqual(t, insta360StackLeft, filepath.Base(left.FileName))
	assert.Equal(t, left.PhotoID, right.PhotoID)
	assert.Equal(t, left.PhotoID, proxy.PhotoID)

	// The right lens and proxy get no preview of their own.
	for _, member := range []entity.File{right, proxy} {
		assert.NoFileExists(t, filepath.Join(cfg.SidecarPath(), member.FileName+".jpg"))
	}

	preview := func() (result entity.File) {
		t.Helper()
		require.NoError(t, entity.UnscopedDb().First(&result, "file_root = ? AND file_name = ?", entity.RootSidecar, left.FileName+".jpg").Error)
		return result
	}

	assert.True(t, preview().FilePrimary)
	assert.Equal(t, "equirectangular", preview().FileProjection)
	assert.Equal(t, 640, preview().FileWidth)
	assertInsta360SinglePrimary(t, filepath.Dir(left.FileName))

	// An earlier import left a preview made from the left lens only.
	previewName := filepath.Join(cfg.SidecarPath(), left.FileName+".jpg")
	writeInsta360Photo(t, cfg, previewName, "320x320")
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", preview().ID).
		UpdateColumn("file_projection", "").Error)

	indexInsta360StackFolder(cfg, filepath.Dir(left.FileName), true, false)

	assert.True(t, preview().FilePrimary)
	assert.Equal(t, "equirectangular", preview().FileProjection)
	assert.Equal(t, 640, preview().FileWidth)
}

// writeInsta360Video writes an H.264 clip of the specified size and duration in seconds.
func writeInsta360Video(t *testing.T, cfg *config.Config, fileName, size, seconds string) {
	t.Helper()
	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), "-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size="+size+":rate=30",
		"-t", seconds, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-metadata", "title="+filepath.Base(fileName), "-f", "mp4", fileName).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))
}

// TestIndex_Insta360FirstIndexMetadata verifies that a capture and its proxy video get their own dimensions,
// duration and codec when they are indexed for the first time.
func TestIndex_Insta360FirstIndexMetadata(t *testing.T) {
	const (
		capture = "VID_20240415_213145_00_035.insv"
		proxy   = "LRV_20240415_213145_01_035.lrv"
	)

	folder := "insta360firstindex"
	cfg := newInsta360StackConfig(t, folder, false)

	if !cfg.ExifToolEnabled() {
		t.Skip("ExifTool must be available to read video metadata")
	}

	dir := filepath.Join(cfg.OriginalsPath(), folder)
	writeInsta360Video(t, cfg, filepath.Join(dir, capture), "384x192", "2")
	writeInsta360Video(t, cfg, filepath.Join(dir, proxy), "192x96", "1")

	indexInsta360StackFolder(cfg, folder, false, true)

	// The proxy is indexed with the capture it belongs to.
	owners := insta360StackOwners(t, folder)
	require.Contains(t, owners, capture)
	require.Contains(t, owners, proxy)
	assert.Equal(t, owners[capture].ID, owners[proxy].ID)

	for _, tc := range []struct {
		name          string
		width, height int
		duration      time.Duration
	}{
		{capture, 384, 192, 2 * time.Second},
		{proxy, 192, 96, time.Second},
	} {
		var file entity.File
		require.NoError(t, entity.UnscopedDb().First(&file, "file_name = ?", folder+"/"+tc.name).Error, tc.name)
		assert.Equal(t, tc.width, file.FileWidth, tc.name)
		assert.Equal(t, tc.height, file.FileHeight, tc.name)
		assert.InDelta(t, tc.duration.Seconds(), file.FileDuration.Seconds(), 0.1, tc.name)
		assert.Equal(t, "avc1", file.FileCodec, tc.name)
	}
}

// sameSecondInsta360Previews gives the previews in folder the next whole second as their file and recorded
// time and waits for it, so a preview replaced right away has the time of the one it replaces.
func sameSecondInsta360Previews(t *testing.T, cfg *config.Config, folder string) time.Time {
	t.Helper()

	next := time.Now().Truncate(time.Second).Add(time.Second)

	matches, err := filepath.Glob(filepath.Join(cfg.SidecarPath(), folder, "*.jpg"))
	require.NoError(t, err)

	for _, fileName := range matches {
		require.NoError(t, os.Chtimes(fileName, next, next))
	}

	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).
		Where("file_root = ? AND file_name LIKE ?", entity.RootSidecar, folder+"/%").
		UpdateColumn("mod_time", next.Unix()).Error)

	time.Sleep(time.Until(next))

	return next
}

// backdateInsta360Previews sets the time of the preview images in dir two seconds back.
func backdateInsta360Previews(t *testing.T, dir string) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(dir, "*.jpg"))
	require.NoError(t, err)

	past := time.Now().Add(-2 * time.Second)

	for _, fileName := range matches {
		require.NoError(t, os.Chtimes(fileName, past, past))
	}
}
