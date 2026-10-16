package photoprism

import (
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
		if stored, match := storedFileMatch(mf); match == storedCopy {
			return stored
		}
		return nil
	}

	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")
	require.NoError(t, err)
	require.NotEmpty(t, mediaFile.Hash())

	t.Run("Nil", func(t *testing.T) {
		stored, match := storedFileMatch(nil)
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
		result := storedFileOf(mf)
		require.NotNil(t, result.file)
		assert.Equal(t, stored, result.file.FileName)
		assert.Equal(t, storedCopy, result.match)
		require.NotNil(t, result.info)
		assert.True(t, result.unchanged())
		require.NotNil(t, result.source)
		assert.True(t, result.sourceUnchanged(copied))

		// A source that changed after it was first read and hashed is not a copy, also if only its time changed.
		changed := filepath.Join(t.TempDir(), "changed.jpg")
		require.NoError(t, fs.Copy(storedPath, changed, false))
		mf, err = NewMediaFile(changed)
		require.NoError(t, err)
		require.Equal(t, file.FileHash, mf.Hash())
		later := mf.FileInfo().ModTime().Add(time.Second)
		require.NoError(t, os.Chtimes(changed, later, later))
		assert.Equal(t, storedNone, storedFileOf(mf).match)
		require.NoError(t, os.WriteFile(changed, []byte("new content"), fs.ModeFile))
		assert.Equal(t, storedNone, storedFileOf(mf).match)
	})
	t.Run("ChangedMaybeSame", func(t *testing.T) {
		// A source that changed since it was read, together with the stored file it may be, is still kept.
		name := filepath.Join(filepath.Dir(storedPath), "changed-maybe-"+rnd.Base36(8)+".jpg")
		writeTestJpeg(t, name, 340)
		row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals,
			FileName: fs.RelName(name, cfg.OriginalsPath()), FileHash: fs.Hash(name)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })
		src := filepath.Join(t.TempDir(), "changed-maybe.jpg")
		require.NoError(t, fs.Copy(name, src, false))

		saved := maybeSameFile
		t.Cleanup(func() { maybeSameFile = saved })
		maybeSameFile = func(a, b os.FileInfo) bool { return b.Name() == filepath.Base(name) }

		mf, err := NewMediaFile(src)
		require.NoError(t, err)
		require.Equal(t, row.FileHash, mf.Hash())

		for _, f := range []string{src, name} {
			appendTo, openErr := os.OpenFile(f, os.O_WRONLY|os.O_APPEND, 0) //nolint:gosec // the file is below a test folder
			require.NoError(t, openErr)
			_, err = appendTo.WriteString("new content")
			require.NoError(t, err)
			require.NoError(t, appendTo.Close())
		}

		found, match := storedFileMatch(mf)
		assert.Equal(t, storedMaybeSame, match)
		require.NotNil(t, found)
		assert.Equal(t, row.FileName, found.FileName)
	})
	t.Run("ChangedOriginal", func(t *testing.T) {
		// A source that changed after it was first read is still recognized as the stored file itself.
		name := filepath.Join(filepath.Dir(storedPath), "changed-"+rnd.Base36(8)+".jpg")
		writeTestJpeg(t, name, 332)
		row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals,
			FileName: fs.RelName(name, cfg.OriginalsPath()), FileHash: fs.Hash(name)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })
		link := filepath.Join(importDir, "changed-link.jpg")
		require.NoError(t, os.Link(name, link))
		t.Cleanup(func() { _ = os.Remove(link) })

		mf, err := NewMediaFile(link)
		require.NoError(t, err)
		require.Equal(t, row.FileHash, mf.Hash())
		f, err := os.OpenFile(link, os.O_WRONLY|os.O_APPEND, 0) //nolint:gosec // the file is below a test folder
		require.NoError(t, err)
		_, err = f.WriteString("new content")
		require.NoError(t, err)
		require.NoError(t, f.Close())

		found, match := storedFileMatch(mf)
		assert.Equal(t, storedSame, match)
		require.NotNil(t, found)
		assert.Equal(t, row.FileName, found.FileName)
	})
	t.Run("HardLink", func(t *testing.T) {
		saved := storedHash
		t.Cleanup(func() { storedHash = saved })
		reads := 0
		storedHash = func(name string) string {
			reads++
			return saved(name)
		}

		mf, err := NewMediaFile(linked)
		require.NoError(t, err)
		found, match := storedFileMatch(mf)
		assert.Equal(t, storedSame, match)
		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)
		assert.Zero(t, reads, "the file itself is found without reading the stored file")
	})
	t.Run("CaseVariant", func(t *testing.T) {
		// A file whose path differs from a stored file only in letter case may be that file, e.g. on a network share
		// whose case mode is unknown, so it is not taken for a copy.
		variant := filepath.Join(filepath.Dir(storedPath), strings.ToUpper(filepath.Base(storedPath)))
		require.NoError(t, fs.Copy(storedPath, variant, false))
		t.Cleanup(func() { _ = os.Remove(variant) })

		mf, err := NewMediaFile(variant)
		require.NoError(t, err)
		found, match := storedFileMatch(mf)
		assert.Equal(t, storedMaybeSame, match)
		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)
	})
	t.Run("Rows", func(t *testing.T) {
		// addRow records rel in originals as a file with the given hash.
		addRow := func(rel, hash string) {
			f := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: rel, FileHash: hash}
			require.NoError(t, f.Save())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(f).Error })
		}

		savedHash := storedHash
		t.Cleanup(func() { storedHash = savedHash })
		var reads []string
		storedHash = func(name string) string {
			reads = append(reads, filepath.Base(name))
			return savedHash(name)
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
		found, match := storedFileMatch(mf)
		assert.Equal(t, storedSame, match)
		require.NotNil(t, found)
		assert.Equal(t, name+"/b.jpg", found.FileName)
		assert.Empty(t, reads, "no file is read before every row is checked for the file itself")

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
		reads = nil
		found, match = storedFileMatch(mf)
		assert.Equal(t, storedCopy, match)
		require.NotNil(t, found)
		assert.Equal(t, copyRel, found.FileName)
		assert.Equal(t, []string{"rows-changed.jpg", "rows-copy.jpg"}, reads, "an absent file is not read")

		// A file that may be the file itself counts only while it holds the content.
		saved := maybeSameFile
		t.Cleanup(func() { maybeSameFile = saved })
		maybeName := "rows-changed.jpg"
		maybeSameFile = func(a, b os.FileInfo) bool { return b.Name() == maybeName }

		found, match = storedFileMatch(mf)
		assert.Equal(t, storedCopy, match)
		require.NotNil(t, found)
		assert.Equal(t, copyRel, found.FileName)

		// A later row that may be the file itself wins over a copy too.
		maybeName = "rows-maybe.jpg"
		maybeRel := filepath.Join(filepath.Dir(stored), "rows-maybe.jpg")
		require.NoError(t, fs.Copy(src, filepath.Join(cfg.OriginalsPath(), maybeRel), false))
		addRow(maybeRel, fs.Hash(src))

		found, match = storedFileMatch(mf)
		assert.Equal(t, storedMaybeSame, match)
		require.NotNil(t, found)
		assert.Equal(t, maybeRel, found.FileName)
	})
	t.Run("Unknown", func(t *testing.T) {
		// addRow records rel in originals as a file with the given hash.
		addRow := func(rel, hash string) *entity.File {
			f := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: rel, FileHash: hash}
			require.NoError(t, f.Save())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(f).Error })
			return f
		}

		src := filepath.Join(t.TempDir(), "unknown.jpg")
		writeTestJpeg(t, src, 317)
		copyRel := filepath.Join(filepath.Dir(stored), "unknown-copy.jpg")
		require.NoError(t, fs.Copy(src, filepath.Join(cfg.OriginalsPath(), copyRel), false))
		addRow(copyRel, fs.Hash(src))
		mf, err := NewMediaFile(src)
		require.NoError(t, err)

		found, match := storedFileMatch(mf)
		assert.Equal(t, storedCopy, match)
		require.NotNil(t, found)

		// A stored file that cannot be looked up may be the file itself, so no other file counts as a copy.
		unreadable := addRow(filepath.Join(filepath.Dir(stored), strings.Repeat("x", 300)+".jpg"), fs.Hash(src))

		found, match = storedFileMatch(mf)
		assert.Equal(t, storedMaybeSame, match)
		assert.Nil(t, found)
		require.NoError(t, entity.UnscopedDb().Delete(unreadable).Error)

		// The same applies to a file that may be the file itself and cannot be read.
		maybeRel := filepath.Join(filepath.Dir(stored), "unknown-maybe.jpg")
		require.NoError(t, fs.Copy(src, filepath.Join(cfg.OriginalsPath(), maybeRel), false))
		addRow(maybeRel, fs.Hash(src))
		savedMaybe, savedHash := maybeSameFile, storedHash
		t.Cleanup(func() { maybeSameFile, storedHash = savedMaybe, savedHash })
		maybeSameFile = func(a, b os.FileInfo) bool { return b.Name() == "unknown-maybe.jpg" }
		storedHash = func(name string) string {
			if filepath.Base(name) == "unknown-maybe.jpg" {
				return ""
			}
			return savedHash(name)
		}

		found, match = storedFileMatch(mf)
		assert.Equal(t, storedMaybeSame, match)
		assert.Nil(t, found)
	})
	t.Run("Sidecar", func(t *testing.T) {
		// A sidecar with the same content is not a copy in the library, while a sidecar that is the file itself is.
		dir := filepath.Join(cfg.SidecarPath(), "zz-sidecar-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		src := filepath.Join(importDir, "sidecar.jpg")
		writeTestJpeg(t, src, 318)
		t.Cleanup(func() { _ = os.Remove(src) })
		require.NoError(t, fs.Copy(src, filepath.Join(dir, "copy.jpg"), false))

		row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootSidecar,
			FileName: filepath.Join(filepath.Base(dir), "copy.jpg"), FileHash: fs.Hash(src)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })

		mf, err := NewMediaFile(src)
		require.NoError(t, err)
		found, match := storedFileMatch(mf)
		assert.Equal(t, storedNone, match)
		assert.Nil(t, found)

		// The destination name still counts the sidecar, also with the files looked up for the group, so a file
		// identical to a preview is not imported as a second picture of it.
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		_, destErr := imp.destinationFilename(mf, mf, "", storedFileOf(mf))
		assert.ErrorContains(t, destErr, "is already indexed as")
		assert.NotErrorIs(t, destErr, errDestinationExists, "a sidecar is not compared as a file under the destination name")
		_, destErr = imp.DestinationFilename(mf, mf, "")
		assert.ErrorContains(t, destErr, "is already indexed as")

		require.NoError(t, os.Remove(filepath.Join(dir, "copy.jpg")))
		require.NoError(t, os.Link(src, filepath.Join(dir, "copy.jpg")))
		found, match = storedFileMatch(mf)
		assert.Equal(t, storedSame, match)
		require.NotNil(t, found)
		assert.Equal(t, row.FileName, found.FileName)
	})
	t.Run("MaybeSame", func(t *testing.T) {
		// A copy with the same size, time and inode number on another device may be the same file.
		stubMaybeSame(t)

		mf, err := NewMediaFile(copied)
		require.NoError(t, err)
		found, match := storedFileMatch(mf)
		assert.Equal(t, storedMaybeSame, match)
		require.NotNil(t, found)
		assert.Equal(t, stored, found.FileName)
	})
	t.Run("MoveImportMaybeSame", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		stubMaybeSame(t)

		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-maybe-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })

		// A main file that may be the original is skipped with its sidecar, and another file is kept.
		for _, name := range []string{"main.jpg", "pair.png"} {
			require.NoError(t, fs.Copy(storedPath, filepath.Join(dir, name), false))
		}

		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.xmp"), []byte(testXmp), fs.ModeFile))
		pair := filepath.Join(dir, "pair.jpg")
		writeTestJpeg(t, pair, 314)
		pairHash := fs.Hash(pair)
		t.Cleanup(func() { cleanupImportedHash(t, pairHash) })

		hook := captureLog(t)
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		opt := ImportOptionsMove(dir, "")
		opt.Albums = []string{testAlbum(t)}

		_, err := imp.Run(opt)
		require.NoError(t, err)

		for _, name := range []string{"main.jpg", "main.xmp", "pair.png"} {
			assert.FileExists(t, filepath.Join(dir, name))
		}

		assert.NoFileExists(t, pair)
		infos := loggedMessages(hook, logrus.InfoLevel)
		assert.Equal(t, 1, countPrefix(infos, "import: skipped main.jpg and its related files, as it may be the same file as the original"))
		assert.Equal(t, 1, countPrefix(infos, "import: kept pair.png, as it may be the same file as the original"))

		// The picture of the stored file that pair.png may be is added to the album, as for a copy.
		assert.Contains(t, albumPhotoUIDs(t, opt.Albums[0]), photo.PhotoUID)
	})
	t.Run("Group", func(t *testing.T) {
		mf := func(name string) *MediaFile {
			f, err := NewMediaFile(name)
			require.NoError(t, err)
			return f
		}

		xmp := filepath.Join(importDir, "linked.xmp")
		require.NoError(t, os.WriteFile(xmp, []byte(strings.Replace(testXmp, "Pair", "Group "+rnd.Base36(8), 1)), fs.ModeFile))
		t.Cleanup(func() { _ = os.Remove(xmp) })

		// A group is original if its files other than sidecars are, and a sidecar has a result only if it is or may
		// be a stored file itself.
		results, originals := storedFilesOf(RelatedFiles{Main: mf(linked), Files: MediaFiles{mf(linked), mf(xmp)}})
		assert.True(t, originals)
		assert.Len(t, results, 1)
		assert.Equal(t, storedSame, results[linked].match)
		assert.NotContains(t, results, xmp)

		results, originals = storedFilesOf(RelatedFiles{Main: mf(linked), Files: MediaFiles{mf(linked), mf(copied)}})
		assert.False(t, originals)
		assert.Equal(t, storedCopy, results[copied].match)

		// Sidecars are only compared next to a file that is or may be a stored file itself.
		results, originals = storedFilesOf(RelatedFiles{Main: mf(copied), Files: MediaFiles{mf(copied), mf(xmp)}})
		assert.False(t, originals)
		assert.Equal(t, storedCopy, results[copied].match)
		assert.NotContains(t, results, xmp)

		_, originals = storedFilesOf(RelatedFiles{})
		assert.False(t, originals)
	})
	t.Run("HashOnce", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// The stored original of a duplicate is read once per group, also if it has a new sidecar.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-once-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		require.NoError(t, fs.Copy(storedPath, filepath.Join(dir, "dup.jpg"), false))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "dup.xmp"), []byte(testXmp), fs.ModeFile))
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		cleanupDestination(t, imp, filepath.Join(dir, "dup.jpg"), filepath.Join(dir, "dup.xmp"))

		saved := storedHash
		t.Cleanup(func() { storedHash = saved })
		reads := 0
		storedHash = func(name string) string {
			if name == storedPath {
				reads++
			}
			return saved(name)
		}

		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.NoFileExists(t, filepath.Join(dir, "dup.jpg"))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: deleted dup.jpg (already exists)"))
		assert.Equal(t, 1, reads)

		// The new sidecar of the duplicate stays in the import folder, as no file is indexed that it could belong to.
		assert.FileExists(t, filepath.Join(dir, "dup.xmp"))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: kept dup.xmp, as its main file is already indexed"))
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: added main"))
		assert.Empty(t, loggedMessages(hook, logrus.WarnLevel))
	})
	t.Run("StoredChanged", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A duplicate is kept if the stored file changes after it was compared.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-changed-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		dup := filepath.Join(dir, "changed.jpg")
		require.NoError(t, fs.Copy(storedPath, dup, false))

		info, err := os.Stat(storedPath)
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chtimes(storedPath, info.ModTime(), info.ModTime()) })

		saved := storedHash
		t.Cleanup(func() { storedHash = saved })
		storedHash = func(name string) string {
			h := saved(name)
			if name == storedPath {
				later := info.ModTime().Add(time.Hour)
				require.NoError(t, os.Chtimes(storedPath, later, later))
			}
			return h
		}

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		hook := captureLog(t)

		_, err = imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, dup)
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: kept changed.jpg, as the stored file has changed"))
	})
	t.Run("SourceChanged", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A duplicate is kept if it is rewritten after it was compared, e.g. by a new upload with the same name.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-rewritten-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		dup := filepath.Join(dir, "rewritten.jpg")
		require.NoError(t, fs.Copy(storedPath, dup, false))

		saved := storedHash
		t.Cleanup(func() { storedHash = saved })
		storedHash = func(name string) string {
			h := saved(name)
			if name == storedPath {
				f, err := os.OpenFile(dup, os.O_WRONLY|os.O_APPEND, 0) //nolint:gosec // the file is below a test folder
				require.NoError(t, err)
				_, err = f.WriteString("new content")
				require.NoError(t, err)
				require.NoError(t, f.Close())
			}
			return h
		}

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, dup)
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: kept rewritten.jpg, as it has changed since it was compared"))
	})
	t.Run("SidecarReimport", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A sidecar imported again with its duplicate is removed too, if the same content is at its destination.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-again-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		main := filepath.Join(dir, "again.jpg")
		require.NoError(t, fs.Copy(storedPath, main, false))
		xmp := filepath.Join(dir, "again.xmp")
		require.NoError(t, os.WriteFile(xmp, []byte(strings.Replace(testXmp, "Pair", filepath.Base(dir), 1)), fs.ModeFile))

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		storedXmp := cleanupDestination(t, imp, main, xmp)
		require.NoError(t, fs.Copy(xmp, storedXmp, false))
		row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals,
			FileName: fs.RelName(storedXmp, cfg.OriginalsPath()), FileHash: fs.Hash(xmp)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })
		hook := captureLog(t)

		// The sidecar is compared with the file under its destination name only, not with all files of its content.
		xmpHash := fs.Hash(xmp)
		savedFiles := filesByHash
		t.Cleanup(func() { filesByHash = savedFiles })
		xmpLookups := 0
		filesByHash = func(hash string) (entity.Files, error) {
			if hash == xmpHash {
				xmpLookups++
			}
			return savedFiles(hash)
		}

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.NoFileExists(t, main)
		assert.NoFileExists(t, xmp)
		assert.FileExists(t, storedXmp)
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: deleted again.xmp (already exists)"))
		assert.Zero(t, xmpLookups)
	})
	t.Run("SidecarAlbums", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A sidecar that is a copy of another picture's sidecar does not add that picture to the albums.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-albums-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		main := filepath.Join(dir, "dup.jpg")
		require.NoError(t, fs.Copy(storedPath, main, false))
		png := filepath.Join(dir, "dup.png")
		writeTestPng(t, png, 330)
		pngHash := fs.Hash(png)
		t.Cleanup(func() { cleanupImportedHash(t, pngHash) })
		xmp := filepath.Join(dir, "dup.xmp")
		require.NoError(t, os.WriteFile(xmp, []byte(strings.Replace(testXmp, "Pair", filepath.Base(dir), 1)), fs.ModeFile))

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		cleanupDestination(t, imp, main, png)

		// The other picture's sidecar has the same content and already holds the name the import gives the XMP.
		other := entity.PhotoFixtures.Get("Photo02")
		otherXmp := cleanupDestination(t, imp, main, xmp)
		require.NoError(t, fs.Copy(xmp, otherXmp, false))
		row := &entity.File{PhotoID: other.ID, PhotoUID: other.PhotoUID, FileRoot: entity.RootOriginals,
			FileName: fs.RelName(otherXmp, cfg.OriginalsPath()), FileHash: fs.Hash(xmp)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })

		opt := ImportOptionsMove(dir, "")
		opt.Albums = []string{testAlbum(t)}
		_, err := imp.Run(opt)
		require.NoError(t, err)
		assert.NoFileExists(t, png)

		uids := albumPhotoUIDs(t, opt.Albums[0])
		assert.Contains(t, uids, photo.PhotoUID)
		assert.NotContains(t, uids, other.PhotoUID)
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
		writeTestPng(t, companion, 315)
		companionHash := fs.Hash(companion)
		t.Cleanup(func() { cleanupImportedHash(t, companionHash) })

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		dest := cleanupDestination(t, imp, filepath.Join(dir, "snap.jpg"), companion)
		hook := captureLog(t)
		opt := ImportOptionsMove(dir, "")
		opt.Albums = []string{testAlbum(t)}

		_, err := imp.Run(opt)
		require.NoError(t, err)
		assert.Contains(t, albumPhotoUIDs(t, opt.Albums[0]), photo.PhotoUID, "the picture of the kept file is added to the album")
		assert.FileExists(t, filepath.Join(dir, "snap.jpg"))
		assert.NoFileExists(t, companion)

		// The moved file is indexed, with the name it had in the import folder.
		indexed, err := entity.FirstFileByHash(companionHash)
		require.NoError(t, err)
		require.NotEqual(t, photo.ID, indexed.PhotoID, "the cleanup removes the picture of the companion")
		assert.Equal(t, fs.RelName(dest, cfg.OriginalsPath()), indexed.FileName)
		assert.Equal(t, "snap.png", indexed.OriginalName)

		infos := loggedMessages(hook, logrus.InfoLevel)
		assert.Zero(t, countPrefix(infos, "import: skipped snap.jpg"))
		assert.Equal(t, 1, countPrefix(infos, "import: kept snap.jpg, as it is the same file as the original"))
	})
	t.Run("LinkedMain", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A main file that is an original through a later row stays in place, although the name of the first row is free.
		name := "zz-linked-main-" + rnd.Base36(8)
		target := filepath.Join(filepath.Dir(cfg.OriginalsPath()), name)
		require.NoError(t, fs.MkdirAll(target))
		require.NoError(t, os.Symlink(target, filepath.Join(cfg.OriginalsPath(), name)))
		t.Cleanup(func() { _ = os.Remove(filepath.Join(cfg.OriginalsPath(), name)); _ = os.RemoveAll(target) })
		main := filepath.Join(target, "c.jpg")
		writeTestJpeg(t, main, 319)
		companion := filepath.Join(target, "c.png")
		writeTestPng(t, companion, 320)
		companionHash := fs.Hash(companion)
		t.Cleanup(func() { cleanupImportedHash(t, companionHash) })

		absentRel := name + "-absent/a.jpg"
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(cfg.OriginalsPath(), filepath.Dir(absentRel))) })

		for _, rel := range []string{absentRel, name + "/c.jpg"} {
			row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: rel, FileHash: fs.Hash(main)}
			require.NoError(t, row.Save())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })
		}

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		cleanupDestination(t, imp, main, companion)
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(target, ""))
		require.NoError(t, err)
		assert.FileExists(t, main)
		assert.NoFileExists(t, filepath.Join(cfg.OriginalsPath(), absentRel))
		assert.NoFileExists(t, companion)
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: kept c.jpg, as it is the same file as the original"))

		// The same applies to a main file that may be an original.
		stubMaybeSame(t)
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), name+"-maybe")
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		main = filepath.Join(dir, "d.jpg")
		writeTestJpeg(t, main, 321)
		companion = filepath.Join(dir, "d.png")
		writeTestPng(t, companion, 322)
		maybeHash := fs.Hash(companion)
		t.Cleanup(func() { cleanupImportedHash(t, maybeHash) })
		maybeRel := filepath.Join(filepath.Dir(stored), "d.jpg")
		require.NoError(t, fs.Copy(main, filepath.Join(cfg.OriginalsPath(), maybeRel), false))

		absentRel = name + "-absent/b.jpg"

		for _, rel := range []string{absentRel, maybeRel} {
			row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: rel, FileHash: fs.Hash(main)}
			require.NoError(t, row.Save())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })
		}

		cleanupDestination(t, imp, main, companion)
		hook = captureLog(t)

		_, err = imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, main)
		assert.NoFileExists(t, filepath.Join(cfg.OriginalsPath(), absentRel))
		assert.NoFileExists(t, companion)
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: kept d.jpg, as it may be the same file as the original"))
	})
	t.Run("HeldMainType", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A new file of a type that is not allowed is not placed in originals in place of a main file that is already
		// indexed, whether the main file is the original itself or a copy of it. The JPEG files are reached first.
		video := filepath.Join(filepath.Dir(storedPath), "held-"+rnd.Base36(8)+".mp4")
		require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "gopher-video.mp4"), video, false))
		row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals,
			FileName: fs.RelName(video, cfg.OriginalsPath()), FileHash: fs.Hash(video)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })

		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-type-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		require.NoError(t, os.Link(video, filepath.Join(dir, "same.mp4")))
		writeTestJpeg(t, filepath.Join(dir, "same.jpg"), 326)
		require.NoError(t, fs.Copy(video, filepath.Join(dir, "copy.mp4"), false))
		writeTestJpeg(t, filepath.Join(dir, "copy.jpg"), 327)
		copyXmp := filepath.Join(dir, "copy.xmp")
		require.NoError(t, os.WriteFile(copyXmp, []byte(strings.Replace(testXmp, "Pair", "copy "+filepath.Base(dir), 1)), fs.ModeFile))

		// An allowed new file is placed and indexed in place of the main file, and so is a sidecar.
		png := filepath.Join(dir, "same.png")
		writeTestPng(t, png, 329)
		pngHash := fs.Hash(png)
		t.Cleanup(func() { cleanupImportedHash(t, pngHash) })
		xmp := filepath.Join(dir, "same.xmp")
		require.NoError(t, os.WriteFile(xmp, []byte(strings.Replace(testXmp, "Pair", filepath.Base(dir), 1)), fs.ModeFile))

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		imp.AllowExt = fs.NewExtList("mp4,png")
		cleanupDestination(t, imp, filepath.Join(dir, "same.mp4"), png)
		cleanupDestination(t, imp, filepath.Join(dir, "same.mp4"), xmp)
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "same.mp4"))
		assert.NoFileExists(t, filepath.Join(dir, "copy.mp4"))
		assert.FileExists(t, filepath.Join(dir, "same.jpg"))
		assert.FileExists(t, filepath.Join(dir, "copy.jpg"))
		assert.NoFileExists(t, png)
		assert.NoFileExists(t, xmp)
		assert.FileExists(t, copyXmp, "no file of its group is indexed")
		warnings := loggedMessages(hook, logrus.WarnLevel)
		assert.Equal(t, 1, countPrefix(warnings, "import: skipped same.jpg because the file type is not allowed"))
		assert.Equal(t, 1, countPrefix(warnings, "import: skipped copy.jpg because the file type is not allowed"))

		indexed, err := entity.FirstFileByHash(pngHash)
		require.NoError(t, err)
		assert.Equal(t, "same.png", indexed.OriginalName)
		require.NotEqual(t, photo.ID, indexed.PhotoID)
	})
	t.Run("PreviewInOriginals", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A file identical to a generated preview is kept, also if the sidecar folder is inside originals.
		saved := cfg.Options().SidecarPath
		t.Cleanup(func() { cfg.Options().SidecarPath = saved })
		cfg.Options().SidecarPath = filepath.Join(cfg.OriginalsPath(), ".zz-sidecar-"+rnd.Base36(8))
		t.Cleanup(func() { _ = os.RemoveAll(cfg.Options().SidecarPath) })
		preview := filepath.Join(cfg.SidecarPath(), "2020", "p.jpg")
		writeTestJpeg(t, preview, 345)
		row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootSidecar, FileName: "2020/p.jpg",
			FileHash: fs.Hash(preview)}
		require.NoError(t, row.Save())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })

		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-preview-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		src := filepath.Join(dir, "download.jpg")
		require.NoError(t, fs.Copy(preview, src, false))

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, src)
		assert.FileExists(t, preview)
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: deleted download.jpg"))
	})
	t.Run("ContentPlacedMeanwhile", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A copy whose destination gets the same content meanwhile, e.g. from another worker, logs no error.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-meanwhile-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		src := filepath.Join(dir, "meanwhile.jpg")
		writeTestJpeg(t, src, 347)
		srcHash := fs.Hash(src)
		t.Cleanup(func() { cleanupImportedHash(t, srcHash) })

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		dest := cleanupDestination(t, imp, src, src)

		saved := copyFileContent
		t.Cleanup(func() { copyFileContent = saved })
		copyFileContent = func(dst io.Writer, r io.Reader) (int64, error) {
			require.NoError(t, fs.Copy(src, dest, false))
			return saved(dst, r)
		}

		hook := captureLog(t)
		_, err := imp.Run(ImportOptionsCopy(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, src)
		assert.Equal(t, srcHash, fs.Hash(dest))
		assert.Empty(t, loggedMessages(hook, logrus.ErrorLevel))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.WarnLevel), "import: "+fs.RelName(dest, cfg.OriginalsPath())+" already holds the content"))
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.WarnLevel), "import: could not copy file"))
	})
	t.Run("LinkedDestination", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// A symbolic link under the destination name to a file with the same content is not taken for a copy.
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-link-dest-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		src := filepath.Join(dir, "linked.jpg")
		writeTestJpeg(t, src, 344)
		external := filepath.Join(t.TempDir(), "external.jpg")
		require.NoError(t, fs.Copy(src, external, false))

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		dest := cleanupDestination(t, imp, src, src)
		require.NoError(t, fs.MkdirAll(filepath.Dir(dest)))
		require.NoError(t, os.Symlink(external, dest))
		hook := captureLog(t)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)
		assert.FileExists(t, src)
		assert.True(t, fs.IsSymlink(dest))
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: deleted linked.jpg"))
	})
	t.Run("LinkedSidecar", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		// The indexed sidecar of an original in a folder linked into originals stays there with it, while a new
		// file next to them is imported.
		name := "zz-linked-side-" + rnd.Base36(8)
		target := filepath.Join(filepath.Dir(cfg.OriginalsPath()), name)
		require.NoError(t, fs.MkdirAll(target))
		require.NoError(t, os.Symlink(target, filepath.Join(cfg.OriginalsPath(), name)))
		t.Cleanup(func() { _ = os.Remove(filepath.Join(cfg.OriginalsPath(), name)); _ = os.RemoveAll(target) })
		main := filepath.Join(target, "e.jpg")
		writeTestJpeg(t, main, 324)
		xmp := filepath.Join(target, "e.xmp")
		require.NoError(t, os.WriteFile(xmp, []byte(strings.Replace(testXmp, "Pair", name, 1)), fs.ModeFile))
		companion := filepath.Join(target, "e.png")
		writeTestPng(t, companion, 325)
		companionHash := fs.Hash(companion)
		t.Cleanup(func() { cleanupImportedHash(t, companionHash) })

		for _, f := range []string{main, xmp} {
			row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals,
				FileName: name + "/" + filepath.Base(f), FileHash: fs.Hash(f)}
			require.NoError(t, row.Save())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })
		}

		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		cleanupDestination(t, imp, main, companion)
		cleanupDestination(t, imp, main, xmp)
		hook := captureLog(t)

		// The sidecar is compared with the file of its name next to the original, without looking up its hash.
		xmpHash := fs.Hash(xmp)
		savedFiles := filesByHash
		t.Cleanup(func() { filesByHash = savedFiles })
		xmpLookups := 0
		filesByHash = func(hash string) (entity.Files, error) {
			if hash == xmpHash {
				xmpLookups++
			}
			return savedFiles(hash)
		}

		_, err := imp.Run(ImportOptionsMove(target, ""))
		require.NoError(t, err)
		assert.FileExists(t, main)
		assert.FileExists(t, xmp)
		assert.NoFileExists(t, companion)
		infos := loggedMessages(hook, logrus.InfoLevel)
		assert.Equal(t, 1, countPrefix(infos, "import: kept e.xmp, as it is the same file as the original"))
		assert.Zero(t, xmpLookups)
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
		opt := ImportOptionsMove(filepath.Join(target, "2019"), "")
		opt.Albums = []string{testAlbum(t)}
		_, err := imp.Run(opt)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(target, "2019", "x.jpg"))
		assert.FileExists(t, filepath.Join(target, "2019", "x.xmp"))
		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: skipped x.jpg and its related files"))
		assert.Zero(t, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: moving"))

		// The picture is added to the albums selected for the import, as for any duplicate.
		assert.Equal(t, []string{photo.PhotoUID}, albumPhotoUIDs(t, opt.Albums[0]))
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

// TestSameFileInfo checks that a file counts as unchanged only while it is the same file with the same size and
// modification time.
func TestSameFileInfo(t *testing.T) {
	name := filepath.Join(t.TempDir(), "a.jpg")

	// info writes the file and returns how it was seen.
	info := func(t *testing.T) os.FileInfo {
		writeTestJpeg(t, name, 323)
		result, err := os.Stat(name)
		require.NoError(t, err)
		return result
	}

	t.Run("Unchanged", func(t *testing.T) {
		assert.True(t, sameFileInfo(name, info(t)))
	})
	t.Run("Modified", func(t *testing.T) {
		i := info(t)
		later := i.ModTime().Add(time.Second)
		require.NoError(t, os.Chtimes(name, later, later))
		assert.False(t, sameFileInfo(name, i))
	})
	t.Run("Resized", func(t *testing.T) {
		i := info(t)
		require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))
		require.NoError(t, os.Chtimes(name, i.ModTime(), i.ModTime()))
		assert.False(t, sameFileInfo(name, i))
	})
	t.Run("Replaced", func(t *testing.T) {
		i := info(t)
		other := name + ".tmp"
		require.NoError(t, fs.Copy(name, other, false))
		require.NoError(t, os.Chtimes(other, i.ModTime(), i.ModTime()))
		require.NoError(t, os.Rename(other, name))
		assert.False(t, sameFileInfo(name, i))
	})
	t.Run("Removed", func(t *testing.T) {
		i := info(t)
		require.NoError(t, os.Remove(name))
		assert.False(t, sameFileInfo(name, i))
	})
}

func TestStoredResult_Unchanged(t *testing.T) {
	cfg := config.TestConfig()
	dir := "zz-unchanged-" + rnd.Base36(8)
	name := filepath.Join(cfg.OriginalsPath(), dir, "a.jpg")
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(name)) })
	writeTestJpeg(t, name, 323)
	info, err := os.Stat(name)
	require.NoError(t, err)
	r := storedResult{file: &entity.File{FileRoot: entity.RootOriginals, FileName: dir + "/a.jpg"}, match: storedCopy, info: info}

	t.Run("Unchanged", func(t *testing.T) {
		assert.True(t, r.unchanged())
	})
	t.Run("NoCopy", func(t *testing.T) {
		assert.False(t, storedResult{}.unchanged())
		assert.False(t, storedResult{file: r.file}.unchanged())
	})
	t.Run("Changed", func(t *testing.T) {
		later := info.ModTime().Add(time.Second)
		require.NoError(t, os.Chtimes(name, later, later))
		assert.False(t, r.unchanged())
	})
}

func TestStoredResult_SourceUnchanged(t *testing.T) {
	name := filepath.Join(t.TempDir(), "a.jpg")
	writeTestJpeg(t, name, 323)
	info, err := os.Stat(name)
	require.NoError(t, err)
	r := storedResult{match: storedCopy, source: info}

	t.Run("Unchanged", func(t *testing.T) {
		assert.True(t, r.sourceUnchanged(name))
	})
	t.Run("NoSource", func(t *testing.T) {
		assert.False(t, storedResult{}.sourceUnchanged(name))
	})
	t.Run("Changed", func(t *testing.T) {
		require.NoError(t, os.WriteFile(name, []byte("new content"), fs.ModeFile))
		assert.False(t, r.sourceUnchanged(name))
	})
}

func TestMainTypeRefused(t *testing.T) {
	dir := t.TempDir()
	jpg := filepath.Join(dir, "a.jpg")
	writeTestJpeg(t, jpg, 328)
	unknown := filepath.Join(dir, "a.unknown")
	require.NoError(t, os.WriteFile(unknown, []byte("x"), fs.ModeFile))

	mf := func(name string) *MediaFile {
		f, err := NewMediaFile(name)
		require.NoError(t, err)
		return f
	}

	t.Run("Allowed", func(t *testing.T) {
		assert.Empty(t, mainTypeRefused(mf(jpg), fs.NewExtList("")))
		assert.Empty(t, mainTypeRefused(mf(jpg), fs.NewExtList("jpg,png")))
	})
	t.Run("NotAllowed", func(t *testing.T) {
		assert.Equal(t, "because the file type is not allowed", mainTypeRefused(mf(jpg), fs.NewExtList("png")))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "due to unknown or unsupported extension", mainTypeRefused(mf(unknown), fs.NewExtList("")))
	})
}

func TestSameInfo(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.jpg"), filepath.Join(dir, "b.jpg")
	writeTestJpeg(t, a, 341)
	require.NoError(t, fs.Copy(a, b, false))
	modTime := time.Date(2019, time.December, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(a, modTime, modTime))
	require.NoError(t, os.Chtimes(b, modTime, modTime))

	stat := func(name string) os.FileInfo {
		info, err := os.Stat(name)
		require.NoError(t, err)
		return info
	}

	t.Run("Same", func(t *testing.T) {
		assert.True(t, sameInfo(stat(a), stat(a)))
	})
	t.Run("OtherFile", func(t *testing.T) {
		assert.False(t, sameInfo(stat(a), stat(b)))
	})
	t.Run("Changed", func(t *testing.T) {
		before := stat(a)
		later := modTime.Add(time.Millisecond)
		require.NoError(t, os.Chtimes(a, later, later))
		assert.False(t, sameInfo(before, stat(a)))
	})
}

func TestStoredMatchOf(t *testing.T) {
	cfg := config.TestConfig()
	dir := "zz-match-of-" + rnd.Base36(8)
	stored := filepath.Join(cfg.OriginalsPath(), dir, "a.jpg")
	writeTestJpeg(t, stored, 342)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(stored)) })
	copied := filepath.Join(t.TempDir(), "a.jpg")
	require.NoError(t, fs.Copy(stored, copied, false))
	mf, err := NewMediaFile(copied)
	require.NoError(t, err)
	row := entity.File{FileRoot: entity.RootOriginals, FileName: dir + "/a.jpg"}

	t.Run("Copy", func(t *testing.T) {
		result := storedMatchOf(mf, mf.Hash(), entity.Files{row})
		assert.Equal(t, storedCopy, result.match)
		assert.False(t, result.queried, "the rows are not recorded")
	})
	t.Run("NoRows", func(t *testing.T) {
		assert.Equal(t, storedNone, storedMatchOf(mf, mf.Hash(), nil).match)
	})
	t.Run("OtherHash", func(t *testing.T) {
		assert.Equal(t, storedNone, storedMatchOf(mf, "other", entity.Files{row}).match)
	})
	t.Run("SourceMissing", func(t *testing.T) {
		missing := &MediaFile{fileName: filepath.Join(t.TempDir(), "missing.jpg")}
		assert.Equal(t, storedNone, storedMatchOf(missing, mf.Hash(), entity.Files{row}).match)
	})
}

func TestImport_DestinationFilenameRows(t *testing.T) {
	cfg := config.TestConfig()
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
	dir := "zz-dest-rows-" + rnd.Base36(8)
	present := filepath.Join(cfg.OriginalsPath(), dir, "present.jpg")
	writeTestJpeg(t, present, 343)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(present)) })
	src := filepath.Join(t.TempDir(), "src.jpg")
	require.NoError(t, fs.Copy(present, src, false))
	mf, err := NewMediaFile(src)
	require.NoError(t, err)

	lookups := 0
	saved := firstFileByHash
	t.Cleanup(func() { firstFileByHash = saved })
	firstFileByHash = func(hash string) (entity.File, error) {
		lookups++
		return saved(hash)
	}

	t.Run("FirstRow", func(t *testing.T) {
		// The oldest row decides, as with FirstFileByHash: a free recorded name is restored.
		rows := entity.Files{{FileRoot: entity.RootOriginals, FileName: dir + "/absent.jpg"}, {FileRoot: entity.RootOriginals, FileName: dir + "/present.jpg"}}
		dest, destErr := imp.destinationFilename(mf, mf, "", storedResult{rows: rows, queried: true})
		require.NoError(t, destErr)
		assert.Equal(t, filepath.Join(cfg.OriginalsPath(), dir, "absent.jpg"), dest)

		_, destErr = imp.destinationFilename(mf, mf, "", storedResult{rows: rows[1:], queried: true})
		assert.ErrorContains(t, destErr, "is already indexed as")
		assert.Zero(t, lookups)
	})
	t.Run("NoRows", func(t *testing.T) {
		// Queried without rows means no lookup, while results that were not queried fall back to one.
		dest, destErr := imp.destinationFilename(mf, mf, "", storedResult{queried: true})
		require.NoError(t, destErr)
		assert.Zero(t, lookups)

		// A file with the same content under the name is reported as such.
		require.NoError(t, fs.Copy(src, dest, false))
		t.Cleanup(func() { _ = os.Remove(dest) })
		taken, destErr := imp.destinationFilename(mf, mf, "", storedResult{queried: true})
		assert.ErrorIs(t, destErr, errDestinationExists)
		assert.Equal(t, dest, taken)

		_, _ = imp.destinationFilename(mf, mf, "", storedResult{})
		assert.Equal(t, 1, lookups)
	})
	t.Run("NoHash", func(t *testing.T) {
		lookups = 0
		gone := filepath.Join(t.TempDir(), "gone.jpg")
		require.NoError(t, os.WriteFile(gone, []byte("x"), fs.ModeFile))
		unreadable, newErr := NewMediaFile(gone)
		require.NoError(t, newErr)
		require.NoError(t, os.Remove(gone))
		require.Empty(t, unreadable.Hash())

		_, _ = imp.destinationFilename(unreadable, unreadable, "", storedResult{})
		assert.Zero(t, lookups)
	})
}

func TestStoredFileAt(t *testing.T) {
	cfg := config.TestConfig()
	photo := entity.PhotoFixtures.Get("Photo01")
	dir := "zz-stored-at-" + rnd.Base36(8)
	stored := filepath.Join(cfg.OriginalsPath(), dir, "a.jpg")
	writeTestJpeg(t, stored, 338)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(stored)) })
	row := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: dir + "/a.jpg",
		FileHash: fs.Hash(stored)}
	require.NoError(t, row.Save())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(row).Error })

	copied := filepath.Join(t.TempDir(), "a.jpg")
	require.NoError(t, fs.Copy(stored, copied, false))
	mf, err := NewMediaFile(copied)
	require.NoError(t, err)

	t.Run("Copy", func(t *testing.T) {
		result := storedFileAt(mf, entity.RootOriginals, dir+"/a.jpg")
		assert.Equal(t, storedCopy, result.match)
		require.NotNil(t, result.file)
		assert.Equal(t, row.ID, result.file.ID)
		assert.False(t, result.queried)
	})
	t.Run("Same", func(t *testing.T) {
		same, sameErr := NewMediaFile(stored)
		require.NoError(t, sameErr)
		assert.Equal(t, storedSame, storedFileAt(same, entity.RootOriginals, dir+"/a.jpg").match)
	})
	t.Run("NotIndexed", func(t *testing.T) {
		// A file under the name with the same content is a copy, also if it is not indexed yet.
		require.NoError(t, fs.Copy(stored, filepath.Join(filepath.Dir(stored), "c.jpg"), false))
		result := storedFileAt(mf, entity.RootOriginals, dir+"/c.jpg")
		assert.Equal(t, storedCopy, result.match)
		require.NotNil(t, result.file)
		assert.Zero(t, result.file.ID)
		assert.True(t, result.unchanged())

		assert.Equal(t, storedNone, storedFileAt(mf, entity.RootOriginals, dir+"/b.jpg").match)
		assert.Equal(t, storedNone, storedFileAt(mf, entity.RootOriginals, "").match)
		assert.Equal(t, storedNone, storedFileAt(nil, entity.RootOriginals, dir+"/a.jpg").match)
	})
}

func TestAlbumStored(t *testing.T) {
	dir := t.TempDir()
	jpg, xmp := filepath.Join(dir, "a.jpg"), filepath.Join(dir, "a.xmp")
	writeTestJpeg(t, jpg, 331)
	require.NoError(t, os.WriteFile(xmp, []byte(testXmp), fs.ModeFile))
	jpgFile, err := NewMediaFile(jpg)
	require.NoError(t, err)
	xmpFile, err := NewMediaFile(xmp)
	require.NoError(t, err)
	row := &entity.File{PhotoUID: "ps6sg6be2lvl0yh8"}

	t.Run("Original", func(t *testing.T) {
		for _, f := range []*MediaFile{jpgFile, xmpFile} {
			assert.True(t, albumStored(f, storedResult{file: row, match: storedSame}))
			assert.True(t, albumStored(f, storedResult{file: row, match: storedMaybeSame}))
		}
	})
	t.Run("Copy", func(t *testing.T) {
		assert.True(t, albumStored(jpgFile, storedResult{file: row, match: storedCopy}))
		assert.False(t, albumStored(xmpFile, storedResult{file: row, match: storedCopy}))
	})
	t.Run("None", func(t *testing.T) {
		assert.False(t, albumStored(jpgFile, storedResult{}))
		assert.False(t, albumStored(jpgFile, storedResult{match: storedMaybeSame}))
		assert.False(t, albumStored(jpgFile, storedResult{file: row, match: storedNone}))
	})
}

func TestStoredPhotoUIDs(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, storedPhotoUIDs(nil))
	})
	t.Run("SortedAndUnique", func(t *testing.T) {
		results := map[string]storedResult{
			"a.jpg": {file: &entity.File{PhotoUID: "ps6sg6be2lvl0yh9"}, match: storedSame},
			"a.xmp": {file: &entity.File{PhotoUID: "ps6sg6be2lvl0yh9"}, match: storedSame},
			"b.jpg": {file: &entity.File{PhotoUID: "ps6sg6be2lvl0yh8"}, match: storedMaybeSame},
			"c.jpg": {match: storedMaybeSame},
			"d.jpg": {file: &entity.File{}, match: storedSame},
			"e.xmp": {file: &entity.File{PhotoUID: "ps6sg6be2lvl0yh7"}, match: storedCopy},
		}
		assert.Equal(t, []string{"ps6sg6be2lvl0yh8", "ps6sg6be2lvl0yh9"}, storedPhotoUIDs(results))
	})
}

// testAlbum returns a new album title and removes the album with that title, if any, when the test ends.
func testAlbum(t *testing.T) string {
	title := "zz-import-" + rnd.Base36(8)

	t.Cleanup(func() {
		var albums entity.Albums
		_ = entity.UnscopedDb().Where("album_title = ?", title).Find(&albums).Error

		for i := range albums {
			_ = entity.UnscopedDb().Delete(&entity.PhotoAlbum{}, "album_uid = ?", albums[i].AlbumUID).Error
			_ = entity.UnscopedDb().Delete(&albums[i]).Error
		}
	})

	return title
}

// albumPhotoUIDs returns the pictures in the album with the given title.
func albumPhotoUIDs(t *testing.T, title string) (uids []string) {
	t.Helper()

	var album entity.Album
	require.NoError(t, entity.UnscopedDb().Where("album_title = ?", title).First(&album).Error)
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("album_uid = ?", album.AlbumUID).Pluck("photo_uid", &uids).Error)

	return uids
}

// storedFileMatch returns the stored copy and the match that storedFileOf reports for a file.
func storedFileMatch(mf *MediaFile) (*entity.File, storedMatch) {
	r := storedFileOf(mf)
	return r.file, r.match
}

// stubMaybeSame reports every file on another device as possibly the same file, as two mounts of one network
// share are.
func stubMaybeSame(t *testing.T) {
	saved := maybeSameFile
	t.Cleanup(func() { maybeSameFile = saved })
	maybeSameFile = func(a, b os.FileInfo) bool { return true }
}

// cleanupDestination removes the file that importing other with main places in originals, and returns its name.
func cleanupDestination(t *testing.T, imp *Import, main, other string) string {
	t.Helper()

	mainFile, err := NewMediaFile(main)
	require.NoError(t, err)
	otherFile, err := NewMediaFile(other)
	require.NoError(t, err)
	dest, err := imp.DestinationFilename(mainFile, otherFile, "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(dest) })

	return dest
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

// TestImport_HashLookups checks that the import looks up the hash of each new media file once, and that two groups
// with the same content, imported by two workers, give one picture.
func TestImport_HashLookups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()

	// countLookups counts the lookups by hash, and by name under the key "name", and calls wait after each lookup of
	// the files with a hash.
	countLookups := func(t *testing.T, wait func(hash string, files entity.Files)) map[string]int {
		var mu sync.Mutex
		counts := make(map[string]int)
		savedFiles, savedFirst, savedName := filesByHash, firstFileByHash, filesByName
		t.Cleanup(func() { filesByHash, firstFileByHash, filesByName = savedFiles, savedFirst, savedName })

		count := func(hash string) {
			mu.Lock()
			counts[hash]++
			mu.Unlock()
		}

		filesByHash = func(hash string) (entity.Files, error) {
			count(hash)
			files, err := savedFiles(hash)
			wait(hash, files)
			return files, err
		}
		filesByName = func(fileRoot, fileName string) (entity.Files, error) {
			count("name")
			return savedName(fileRoot, fileName)
		}
		firstFileByHash = func(hash string) (entity.File, error) {
			count(hash)
			return savedFirst(hash)
		}

		return counts
	}

	t.Run("OnePerNewFile", func(t *testing.T) {
		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-lookups-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })

		names := map[string]int{"a.jpg": 334, "a.png": 335, "b.jpg": 336}
		hashes := make([]string, 0, len(names))

		for name, seed := range names {
			if strings.HasSuffix(name, ".png") {
				writeTestPng(t, filepath.Join(dir, name), seed)
			} else {
				writeTestJpeg(t, filepath.Join(dir, name), seed)
			}

			hash := fs.Hash(filepath.Join(dir, name))
			hashes = append(hashes, hash)
			t.Cleanup(func() { cleanupImportedHash(t, hash) })
		}

		for _, name := range []string{"a.xmp", "b.xmp"} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(strings.Replace(testXmp, "Pair", name+" "+filepath.Base(dir), 1)), fs.ModeFile))
		}

		counts := countLookups(t, func(string, entity.Files) {})
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)

		total := 0

		for _, c := range counts {
			total += c
		}

		for _, hash := range hashes {
			assert.Equal(t, 1, counts[hash], hash)
		}

		assert.Equal(t, len(names), total, "sidecars are not looked up")
		assert.Zero(t, counts["name"])
	})
	// importIdentical imports two copies of one picture in two groups with two workers, calling wait for each lookup
	// of its hash, checks that one copy ends up in originals, and returns how many stay in the import folder.
	importIdentical := func(t *testing.T, seed int, wait func(t *testing.T, arrival int32, hash string, files entity.Files)) int {
		saved := cfg.Options().IndexWorkers
		t.Cleanup(func() { cfg.Options().IndexWorkers = saved })
		cfg.Options().IndexWorkers = "2"

		if cfg.IndexWorkers() < 2 {
			t.Skip("skipping test with fewer than two index workers.")
		}

		dir := filepath.Join(filepath.Dir(cfg.OriginalsPath()), "import-twice-"+rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(filepath.Join(dir, "one")))
		require.NoError(t, fs.MkdirAll(filepath.Join(dir, "two")))
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		first, second := filepath.Join(dir, "one", "x.jpg"), filepath.Join(dir, "two", "y.jpg")
		writeTestJpeg(t, first, seed)
		require.NoError(t, fs.Copy(first, second, false))
		modTime := time.Date(2019, time.December, 1, 12, 0, 0, 0, time.UTC)

		for _, name := range []string{first, second} {
			require.NoError(t, os.Chtimes(name, modTime, modTime))
		}

		hash := fs.Hash(first)
		t.Cleanup(func() {
			cleanupImportedHash(t, hash)
			_ = entity.UnscopedDb().Delete(&entity.Duplicate{}, "file_hash = ?", hash).Error
		})

		// Both copies get the same destination name, which is all the second worker has to go by.
		convert := NewConvert(cfg)
		imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
		dest := cleanupDestination(t, imp, first, first)
		require.Equal(t, dest, cleanupDestination(t, imp, second, second))

		var arrivals atomic.Int32
		countLookups(t, func(h string, files entity.Files) {
			if h == hash {
				wait(t, arrivals.Add(1), hash, files)
			}
		})

		hook := captureLog(t)
		_, err := imp.Run(ImportOptionsMove(dir, ""))
		require.NoError(t, err)

		var copies []string
		require.NoError(t, filepath.WalkDir(filepath.Dir(dest), func(name string, d iofs.DirEntry, walkErr error) error {
			if walkErr == nil && !d.IsDir() && fs.Hash(name) == hash {
				copies = append(copies, name)
			}
			return walkErr
		}))
		assert.Equal(t, []string{dest}, copies)

		remaining := 0

		for _, name := range []string{first, second} {
			if fs.FileExists(name) {
				remaining++
			}
		}

		assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.InfoLevel), "import: added main jpg file"))
		assert.Empty(t, loggedMessages(hook, logrus.ErrorLevel))

		return remaining
	}

	t.Run("IdenticalGroups", func(t *testing.T) {
		// Both groups look up the hash before either is moved, so both choose the same destination name.
		var arrived sync.WaitGroup
		arrived.Add(2)
		release := make(chan struct{})
		var once sync.Once
		var timedOut atomic.Bool
		go func() {
			arrived.Wait()
			once.Do(func() { close(release) })
		}()
		timer := time.AfterFunc(10*time.Second, func() {
			once.Do(func() { timedOut.Store(true); close(release) })
		})
		t.Cleanup(func() { timer.Stop() })

		remaining := importIdentical(t, 337, func(t *testing.T, arrival int32, hash string, files entity.Files) {
			if arrival <= 2 {
				arrived.Done()
				<-release
			}
		})

		assert.False(t, timedOut.Load(), "both workers looked up the hash before either moved its file")
		assert.LessOrEqual(t, remaining, 1, "one copy was moved")
	})
	t.Run("IdenticalGroupsAfterIndexing", func(t *testing.T) {
		// The first group goes on once the second has looked up the hash, and the second once the first is indexed,
		// so the second relies on the file under the destination name rather than on a row, and removes its copy.
		looked := make(chan struct{})
		var staleRows atomic.Bool

		remaining := importIdentical(t, 339, func(t *testing.T, arrival int32, hash string, files entity.Files) {
			switch arrival {
			case 1:
				select {
				case <-looked:
				case <-time.After(10 * time.Second):
				}
			case 2:
				staleRows.Store(len(files) == 0)
				close(looked)
				assert.Eventually(t, func() bool {
					var count int64
					_ = entity.UnscopedDb().Model(&entity.File{}).Where("file_hash = ? AND file_root = ?", hash, entity.RootOriginals).Count(&count).Error
					return count > 0
				}, 10*time.Second, 50*time.Millisecond)
			}
		})

		assert.True(t, staleRows.Load(), "the second group found no rows")
		assert.Zero(t, remaining)
	})
}
