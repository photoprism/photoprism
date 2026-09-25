package photoprism

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
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

	var args []string

	if fileType := fs.FileType(name); fileType == fs.ImageJpeg || fileType == fs.ImageInsp {
		args = []string{"-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x320",
			"-frames:v", "1", "-metadata", "comment=" + name, "-f", "image2", "-c:v", "mjpeg", filepath.Join(dir, name)}
	} else {
		args = []string{"-y", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x320:rate=30",
			"-t", "1", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-metadata", "title=" + name, "-f", "mp4",
			filepath.Join(dir, name)}
	}

	// #nosec G204 -- arguments are test constants.
	out, err := exec.Command(cfg.FFmpegBin(), args...).CombinedOutput()
	require.NoError(t, err, strings.TrimSpace(string(out)))
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
		assert.Equal(t, "VID_20220625_140410_10_008", owners["VID_20220625_140410_10_008.mp4"].PhotoName)
	})
	t.Run("MovedLens", func(t *testing.T) {
		folder := "insta360controlmoved"
		movedFolder := "insta360controlmovedto"
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
