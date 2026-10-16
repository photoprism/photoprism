package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewImport(t *testing.T) {
	cfg := config.TestConfig()

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	assert.IsType(t, &Import{}, imp)
}

func TestStoredCopyOf(t *testing.T) {
	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")
	require.NoError(t, err)
	require.NotEmpty(t, mediaFile.Hash())

	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, StoredCopyOf(nil))
	})
	t.Run("NoHash", func(t *testing.T) {
		dir := t.TempDir()
		name := filepath.Join(dir, "gone.jpg")
		require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))

		unreadable, newErr := NewMediaFile(name)
		require.NoError(t, newErr)
		require.NoError(t, os.Remove(name))
		require.Empty(t, unreadable.Hash())

		assert.Nil(t, StoredCopyOf(unreadable))
	})
	t.Run("NotIndexed", func(t *testing.T) {
		assert.Nil(t, StoredCopyOf(mediaFile), "no row records this content")
	})
	t.Run("IndexedButNotOnDisk", func(t *testing.T) {
		// A row on its own is not the content. Answering otherwise lets a caller remove the copy it
		// is holding.
		photo := entity.PhotoFixtures.Get("Photo01")
		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: "zz-stored/absent.cr2",
			FileHash: mediaFile.Hash(),
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(file).Error })

		assert.Nil(t, StoredCopyOf(mediaFile))
	})
	t.Run("IndexedButContentChanged", func(t *testing.T) {
		// The row records what the file held when it was read. If the file has changed since, the
		// library no longer holds this content, and answering otherwise removes the only copy.
		photo := entity.PhotoFixtures.Get("Photo01")
		stored := "zz-stored/changed.cr2"
		storedPath := filepath.Join(cfg.OriginalsPath(), stored)

		require.NoError(t, os.MkdirAll(filepath.Dir(storedPath), fs.ModeDir))
		require.NoError(t, os.WriteFile(storedPath, []byte("different content"), fs.ModeFile))

		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: stored,
			FileHash: mediaFile.Hash(),
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(file).Error
			_ = os.RemoveAll(filepath.Dir(storedPath))
		})

		require.NotEqual(t, mediaFile.Hash(), fs.Hash(storedPath), "the file must no longer match its row")

		assert.Nil(t, StoredCopyOf(mediaFile))
	})
	t.Run("IndexedAndOnDisk", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("Photo01")
		stored := "zz-stored/present.cr2"
		storedPath := filepath.Join(cfg.OriginalsPath(), stored)

		require.NoError(t, os.MkdirAll(filepath.Dir(storedPath), fs.ModeDir))
		require.NoError(t, fs.Copy(mediaFile.FileName(), storedPath, false))

		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: stored,
			FileHash: mediaFile.Hash(),
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(file).Error
			_ = os.RemoveAll(filepath.Dir(storedPath))
		})

		found := StoredCopyOf(mediaFile)

		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)

		// The stored file itself, reached through a link or by its own name, is not a copy of it.
		link := filepath.Join(t.TempDir(), "link")
		require.NoError(t, os.Symlink(filepath.Dir(storedPath), link))

		for _, name := range []string{storedPath, filepath.Join(link, "present.cr2")} {
			same, sameErr := NewMediaFile(name)
			require.NoError(t, sameErr)
			assert.Nil(t, StoredCopyOf(same), name)
		}
	})
}

func TestStoredFileOf(t *testing.T) {
	cfg := config.TestConfig()
	photo := entity.PhotoFixtures.Get("Photo01")
	stored := "zz-stored-same/" + rnd.Base36(8) + ".jpg"
	storedPath := filepath.Join(cfg.OriginalsPath(), stored)
	writeTestJpeg(t, storedPath, 310)

	file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: stored,
		FileHash: fs.Hash(storedPath)}
	require.NoError(t, file.Save())
	t.Cleanup(func() {
		_ = entity.UnscopedDb().Delete(file).Error
		_ = os.RemoveAll(filepath.Dir(storedPath))
	})

	// A hard link outside originals, on the same file system, is the same file.
	importDir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-same-"+rnd.Base36(8))
	require.NoError(t, fs.MkdirAll(importDir))
	t.Cleanup(func() { _ = os.RemoveAll(importDir) })
	linked := filepath.Join(importDir, "linked.jpg")
	require.NoError(t, os.Link(storedPath, linked))
	copied := filepath.Join(importDir, "copied.jpg")
	require.NoError(t, fs.Copy(storedPath, copied, false))

	t.Run("Copy", func(t *testing.T) {
		mf, err := NewMediaFile(copied)
		require.NoError(t, err)
		found, same := storedFileOf(mf)
		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)
		assert.False(t, same)
		assert.NotNil(t, StoredCopyOf(mf))
	})
	t.Run("HardLink", func(t *testing.T) {
		mf, err := NewMediaFile(linked)
		require.NoError(t, err)
		found, same := storedFileOf(mf)
		require.NotNil(t, found)
		assert.True(t, same)
		assert.Nil(t, StoredCopyOf(mf))
	})
	t.Run("Nil", func(t *testing.T) {
		found, same := storedFileOf(nil)
		assert.Nil(t, found)
		assert.False(t, same)
	})
	t.Run("MoveImport", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		require.NoError(t, os.Remove(copied))
		require.NoError(t, os.WriteFile(filepath.Join(importDir, "linked.xmp"), []byte(testXmp), fs.ModeFile))

		// A related file that is the same file as an original is kept, while its main file is imported.
		pair := filepath.Join(importDir, "pair.jpg")
		writeTestJpeg(t, pair, 311)
		require.NoError(t, os.Link(storedPath, filepath.Join(importDir, "pair.png")))
		pairHash := fs.Hash(pair)
		t.Cleanup(func() { cleanupImportedHash(t, pairHash) })

		hook := captureLog(t)
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

		_, err := imp.Run(ImportOptionsMove(importDir, ""))
		require.NoError(t, err)
		assert.FileExists(t, linked)
		assert.FileExists(t, filepath.Join(importDir, "linked.xmp"))
		assert.FileExists(t, storedPath)
		assert.FileExists(t, filepath.Join(importDir, "pair.png"))
		assert.NoFileExists(t, pair)

		infos := loggedMessages(hook, logrus.InfoLevel)
		assert.Equal(t, 1, countPrefix(infos, "import: skipped linked.jpg and its related files, as it is the same file as the original"))
		assert.Equal(t, 1, countPrefix(infos, "import: kept pair.png, as it is the same file as the original"))
	})
}

// TestImport_RunLinkedOriginals checks that files in a folder that originals links to, or in a hidden folder in
// originals indexed before, are left in place with their sidecars.
func TestImport_RunLinkedOriginals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	photo := entity.PhotoFixtures.Get("Photo01")
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	// stage writes a picture with a sidecar to dir and records the picture as indexed under rel in originals.
	// Each case uses its own picture, so a sidecar one case moves cannot hide what another does.
	stage := func(t *testing.T, dir, rel string, seed int) {
		writeTestJpeg(t, filepath.Join(dir, "x.jpg"), seed)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "x.xmp"), []byte(testXmp), fs.ModeFile))
		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: rel,
			FileHash: fs.Hash(filepath.Join(dir, "x.jpg"))}
		require.NoError(t, file.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(file).Error })
	}

	t.Run("Link", func(t *testing.T) {
		name := "linked-" + rnd.Base36(8)
		target := filepath.Join(filepath.Dir(cfg.OriginalsPath()), name)
		link := filepath.Join(cfg.OriginalsPath(), name)
		require.NoError(t, fs.MkdirAll(filepath.Join(target, "2019")))
		require.NoError(t, os.Symlink(target, link))
		t.Cleanup(func() { _ = os.Remove(link); _ = os.RemoveAll(target) })
		stage(t, filepath.Join(target, "2019"), name+"/2019/x.jpg", 312)

		hook := captureLog(t)
		_, err := imp.Run(ImportOptionsMove(filepath.Join(target, "2019"), ""))
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(target, "2019", "x.jpg"))
		assert.FileExists(t, filepath.Join(target, "2019", "x.xmp"))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: skipped x.jpg and its related files"))
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: moving"))
	})
	t.Run("HiddenFolder", func(t *testing.T) {
		name := "~Archive-" + rnd.Base36(8)
		dir := filepath.Join(cfg.OriginalsPath(), name, "2019")
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(dir)) })
		stage(t, dir, name+"/2019/x.jpg", 313)

		hook := captureLog(t)
		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "x.jpg"))
		assert.FileExists(t, filepath.Join(dir, "x.xmp"))
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: moving"))
	})
}

func TestImport_DestinationFilename(t *testing.T) {
	cfg := config.TestConfig()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)

	rawFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("NoBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
	t.Run("WithBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "users/guest")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg.OriginalsPath()+"/users/guest/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
	t.Run("ARecordedNameHeldByALinkIsNotReportedAsADuplicate", func(t *testing.T) {
		// A duplicate is reported only when the recorded file is there. The worker deletes the file
		// it is importing on that answer, so a name that resolves to nothing takes the search
		// instead.
		hash := rawFile.Hash()
		require.NotEmpty(t, hash)

		recorded := "zz-linked/recorded.cr2"
		recordedPath := filepath.Join(cfg.OriginalsPath(), recorded)

		require.NoError(t, os.MkdirAll(filepath.Dir(recordedPath), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Join(cfg.OriginalsPath(), "zz-linked", "gone.cr2"), recordedPath))

		photo := entity.PhotoFixtures.Get("Photo01")

		file := &entity.File{
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileRoot: entity.RootOriginals,
			FileName: recorded,
			FileHash: hash,
		}
		require.NoError(t, file.Save())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(file).Error
			_ = os.RemoveAll(filepath.Dir(recordedPath))
		})

		require.False(t, fs.FileExists(recordedPath), "the recorded name must resolve to nothing")

		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		require.NoError(t, err, "a name that resolves to nothing is not a duplicate")
		assert.NotEqual(t, recordedPath, fileName, "and must not be handed back as the destination")
	})
	t.Run("AnUnreadableSourceDoesNotMatchALinkWithNoTarget", func(t *testing.T) {
		// Both sides hash to the empty string when they cannot be read, so a hash is compared only
		// once there is one on each side.
		dir := t.TempDir()

		name := filepath.Join(dir, "unreadable.jpg")
		require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))

		unreadable, err := NewMediaFile(name)
		require.NoError(t, err)

		// The hash is computed once and kept, so the source goes before anything asks for it.
		require.NoError(t, os.Remove(name))
		require.Empty(t, unreadable.Hash(), "the source must not be readable, which is the case at issue")

		// Learn the name the search settles on before anything holds it.
		free, err := imp.DestinationFilename(unreadable, unreadable, "")
		require.NoError(t, err)

		require.NoError(t, os.MkdirAll(filepath.Dir(free), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Join(dir, "no-such-target.jpg"), free))

		t.Cleanup(func() { _ = os.Remove(free) })

		require.Empty(t, fs.Hash(free), "and the name must be held by something that cannot be read either")

		fileName, err := imp.DestinationFilename(unreadable, unreadable, "")

		require.NoError(t, err, "two unreadable names are not the same file")
		assert.NotEqual(t, free, fileName, "the held name must be stepped over")
	})
	t.Run("StepsOverASymbolicLink", func(t *testing.T) {
		// The search reads the name rather than what it resolves to, because the move that follows
		// refuses a link.
		taken := cfg.OriginalsPath() + "/2019/07/20190705_153230_C167C6FD.cr2"

		require.NoError(t, os.MkdirAll(filepath.Dir(taken), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Join(cfg.OriginalsPath(), "no-such-target.cr2"), taken))

		t.Cleanup(func() { _ = os.Remove(taken) })

		require.False(t, fs.FileExists(taken), "the link must not resolve, which is the case at issue")

		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		require.NoError(t, err)
		assert.NotEqual(t, taken, fileName, "the name a link holds must not be handed back")
		assert.Equal(t, cfg.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.00001.cr2", fileName)
	})
}

func TestImport_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)

	opt := ImportOptionsMove(cfg.ImportPath(), "")

	imp.Start(opt)
}

// TestImport_StartMissingPath verifies that the notification for a missing folder names it relative to import.
func TestImport_StartMissingPath(t *testing.T) {
	cfg := config.TestConfig()
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	s := event.Subscribe("notify.error")
	defer event.Unsubscribe(s)

	imp.Start(ImportOptionsCopy(filepath.Join(cfg.ImportPath(), "e2e-missing/folder"), ""))

	assertMissingFolderNotice(t, s.Receiver, "import: directory", "e2e-missing/folder", cfg.ImportPath())
}

// cleanupImportedHash removes the picture a test imported with a file of the given hash, and its files in originals.
func cleanupImportedHash(t *testing.T, hash string) {
	t.Helper()

	file, err := entity.FirstFileByHash(hash)

	if err != nil {
		return
	}

	var files entity.Files
	_ = entity.UnscopedDb().Where("photo_id = ?", file.PhotoID).Find(&files).Error

	for _, f := range files {
		if f.FileRoot == entity.RootOriginals {
			_ = os.Remove(FileName(f.FileRoot, f.FileName))
		}
	}

	_ = entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID).Error
	_ = entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID).Error
	_ = entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID).Error
}
