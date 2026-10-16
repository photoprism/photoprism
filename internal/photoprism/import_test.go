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

// TestStoredFileOf_Content checks that an indexed file counts as a copy only while it holds the same content.
func TestStoredFileOf_Content(t *testing.T) {
	// storedCopyOf returns the indexed file that is a copy of a file, if there is one.
	storedCopyOf := func(mf *MediaFile) *entity.File {
		stored, _ := storedFileOf(mf)
		return stored
	}

	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")
	require.NoError(t, err)
	require.NotEmpty(t, mediaFile.Hash())

	t.Run("Nil", func(t *testing.T) {
		stored, match := storedFileOf(nil)
		assert.Nil(t, stored)
		assert.Equal(t, storedNone, match)
	})
	t.Run("NoHash", func(t *testing.T) {
		dir := t.TempDir()
		name := filepath.Join(dir, "gone.jpg")
		require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))

		unreadable, newErr := NewMediaFile(name)
		require.NoError(t, newErr)
		require.NoError(t, os.Remove(name))
		require.Empty(t, unreadable.Hash())

		assert.Nil(t, storedCopyOf(unreadable))
	})
	t.Run("NotIndexed", func(t *testing.T) {
		assert.Nil(t, storedCopyOf(mediaFile), "no row records this content")
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

		assert.Nil(t, storedCopyOf(mediaFile))
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

		assert.Nil(t, storedCopyOf(mediaFile))
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

		found := storedCopyOf(mediaFile)

		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)

		// The stored file itself, reached through a link or by its own name, is not a copy of it.
		link := filepath.Join(t.TempDir(), "link")
		require.NoError(t, os.Symlink(filepath.Dir(storedPath), link))

		for _, name := range []string{storedPath, filepath.Join(link, "present.cr2")} {
			same, sameErr := NewMediaFile(name)
			require.NoError(t, sameErr)
			assert.Nil(t, storedCopyOf(same), name)
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
		found, match := storedFileOf(mf)
		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)
		assert.Equal(t, storedCopy, match)
	})
	t.Run("HardLink", func(t *testing.T) {
		mf, err := NewMediaFile(linked)
		require.NoError(t, err)
		found, match := storedFileOf(mf)
		assert.Equal(t, storedSame, match)
		assert.Nil(t, found)
	})
	t.Run("Rows", func(t *testing.T) {
		// addRow records rel in originals as a file with the given hash.
		addRow := func(rel, hash string) {
			f := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: rel, FileHash: hash}
			require.NoError(t, f.Save())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(f).Error })
		}

		// A later row that is the file itself, through a folder linked into originals, wins over a copy.
		name := "zz-link-" + rnd.Base36(8)
		target := filepath.Join(filepath.Dir(cfg.OriginalsPath()), name)
		require.NoError(t, fs.MkdirAll(target))
		require.NoError(t, os.Symlink(target, filepath.Join(cfg.OriginalsPath(), name)))
		t.Cleanup(func() { _ = os.Remove(filepath.Join(cfg.OriginalsPath(), name)); _ = os.RemoveAll(target) })
		require.NoError(t, fs.Copy(storedPath, filepath.Join(target, "b.jpg"), false))
		addRow(name+"/b.jpg", file.FileHash)

		mf, err := NewMediaFile(filepath.Join(target, "b.jpg"))
		require.NoError(t, err)
		found, match := storedFileOf(mf)
		assert.Equal(t, storedSame, match)
		assert.Nil(t, found)

		// A later row whose file still holds the content is a copy, if earlier ones do not.
		src := filepath.Join(t.TempDir(), "rows.jpg")
		writeTestJpeg(t, src, 316)
		copyRel := filepath.Join(filepath.Dir(stored), "rows-copy.jpg")
		changedRel := filepath.Join(filepath.Dir(stored), "rows-changed.jpg")
		require.NoError(t, fs.Copy(src, filepath.Join(cfg.OriginalsPath(), copyRel), false))
		require.NoError(t, os.WriteFile(filepath.Join(cfg.OriginalsPath(), changedRel), []byte("changed"), fs.ModeFile))
		addRow(filepath.Join(filepath.Dir(stored), "rows-absent.jpg"), fs.Hash(src))
		addRow(changedRel, fs.Hash(src))
		addRow(copyRel, fs.Hash(src))

		mf, err = NewMediaFile(src)
		require.NoError(t, err)
		found, match = storedFileOf(mf)
		assert.Equal(t, storedCopy, match)
		require.NotNil(t, found)
		assert.Equal(t, copyRel, found.FileName)

		// A later row that may be the file itself wins over a copy too.
		saved := maybeSameFile
		t.Cleanup(func() { maybeSameFile = saved })
		maybeSameFile = func(a, b os.FileInfo) bool { return b.Name() == "rows-maybe.jpg" }
		maybeRel := filepath.Join(filepath.Dir(stored), "rows-maybe.jpg")
		require.NoError(t, fs.Copy(src, filepath.Join(cfg.OriginalsPath(), maybeRel), false))
		addRow(maybeRel, fs.Hash(src))

		found, match = storedFileOf(mf)
		assert.Equal(t, storedMaybeSame, match)
		assert.Nil(t, found)
	})
	t.Run("MaybeSame", func(t *testing.T) {
		// A copy with the same size, time and inode number on another device may be the same file.
		stubMaybeSame(t)
		info, err := os.Stat(storedPath)
		require.NoError(t, err)
		require.NoError(t, os.Chtimes(copied, info.ModTime(), info.ModTime()))

		mf, err := NewMediaFile(copied)
		require.NoError(t, err)
		found, match := storedFileOf(mf)
		assert.Equal(t, storedMaybeSame, match)
		assert.Nil(t, found)
	})
	t.Run("MoveImportMaybeSame", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		stubMaybeSame(t)
		info, err := os.Stat(storedPath)
		require.NoError(t, err)

		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-maybe-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })

		// A main file that may be the original is skipped with its sidecar, and another file is kept.
		for _, name := range []string{"main.jpg", "pair.png"} {
			require.NoError(t, fs.Copy(storedPath, filepath.Join(dir, name), false))
			require.NoError(t, os.Chtimes(filepath.Join(dir, name), info.ModTime(), info.ModTime()))
		}

		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.xmp"), []byte(testXmp), fs.ModeFile))
		pair := filepath.Join(dir, "pair.jpg")
		writeTestJpeg(t, pair, 314)
		pairHash := fs.Hash(pair)
		t.Cleanup(func() { cleanupImportedHash(t, pairHash) })

		hook := captureLog(t)
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

		_, err = imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)

		for _, name := range []string{"main.jpg", "main.xmp", "pair.png"} {
			assert.FileExists(t, filepath.Join(dir, name))
		}

		assert.NoFileExists(t, pair)
		infos := loggedMessages(hook, logrus.InfoLevel)
		assert.Equal(t, 1, countPrefix(infos, "import: skipped main.jpg and its related files, as it may be the same file as the original"))
		assert.Equal(t, 1, countPrefix(infos, "import: kept pair.png, as it may be the same file as the original"))
	})
	t.Run("Group", func(t *testing.T) {
		mf := func(name string) *MediaFile {
			f, err := NewMediaFile(name)
			require.NoError(t, err)
			return f
		}

		xmp := filepath.Join(importDir, "linked.xmp")
		require.NoError(t, os.WriteFile(xmp, []byte(testXmp), fs.ModeFile))
		t.Cleanup(func() { _ = os.Remove(xmp) })

		// Sidecars are not looked up, and a group is original only if all other files are.
		results, originals := storedFilesOf(RelatedFiles{Main: mf(linked), Files: MediaFiles{mf(linked), mf(xmp)}})
		assert.True(t, originals)
		assert.Len(t, results, 1)
		assert.Equal(t, storedSame, results[linked].match)

		results, originals = storedFilesOf(RelatedFiles{Main: mf(linked), Files: MediaFiles{mf(linked), mf(copied)}})
		assert.False(t, originals)
		assert.Equal(t, storedCopy, results[copied].match)

		_, originals = storedFilesOf(RelatedFiles{})
		assert.False(t, originals)
	})
	t.Run("HashOnce", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// The stored original of a duplicate is read once per import, also if it has a new sidecar.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-once-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		require.NoError(t, fs.Copy(storedPath, filepath.Join(dir, "dup.jpg"), false))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "dup.xmp"), []byte(testXmp), fs.ModeFile))

		saved := storedHash
		t.Cleanup(func() { storedHash = saved })
		reads := 0
		storedHash = func(name string) string {
			if name == storedPath {
				reads++
			}
			return saved(name)
		}

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.NoFileExists(t, filepath.Join(dir, "dup.jpg"))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: deleted dup.jpg (already exists)"))
		assert.Equal(t, 1, reads)
	})
	t.Run("Snapshot", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A hard link to an original next to a new file, e.g. in a snapshot, does not keep the new file back.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-snap-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		require.NoError(t, os.Link(storedPath, filepath.Join(dir, "snap.jpg")))
		companion := filepath.Join(dir, "snap.png")
		writeTestJpeg(t, companion, 315)
		companionHash := fs.Hash(companion)
		t.Cleanup(func() { cleanupImportedHash(t, companionHash) })

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "snap.jpg"))
		assert.NoFileExists(t, companion)

		infos := loggedMessages(hook, logrus.InfoLevel)
		assert.Zero(t, countPrefix(infos, "import: skipped snap.jpg"))
		assert.Equal(t, 1, countPrefix(infos, "import: kept snap.jpg, as it is the same file as the original"))
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
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: skipped x.jpg and its related files"))
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

// stubMaybeSame reports every file on another device as possibly the same file, as two mounts of one network
// share are.
func stubMaybeSame(t *testing.T) {
	saved := maybeSameFile
	t.Cleanup(func() { maybeSameFile = saved })
	maybeSameFile = func(a, b os.FileInfo) bool { return true }
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
