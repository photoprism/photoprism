package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestImport_StartUnportableNames checks that the import leaves files and folders whose names are not portable in place.
func TestImport_StartUnportableNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	src := t.TempDir()

	portable := []string{"ok.jpg", "03 - Março/#1 & Friends [2020] (copy) Tom's.jpg", "pair.jpg", "stack.jpg", "日本\u3000旅行/😀.jpg"}
	skipped := []string{"a|b.jpg", "x:y.jpg", "dir*/inner.jpg", "caf\xe9.jpg", "a~b.jpg", "a$b.jpg", "pair.x|y.xmp", "stack.x|y.gif"}

	for i, name := range portable {
		writeTestJpeg(t, filepath.Join(src, name), 200+i)
	}

	for i, name := range skipped {
		switch filepath.Ext(name) {
		case ".xmp":
			require.NoError(t, os.WriteFile(filepath.Join(src, name), []byte(testXmp), fs.ModeFile))
		case ".gif":
			writeTestGif(t, filepath.Join(src, name))
		default:
			writeTestJpeg(t, filepath.Join(src, name), 220+i)
		}
	}

	// Hidden files are not imported. A move removes them, except files of a known type whose names do not start with
	// a dot, of which media files are reported, and hidden folders, which stay with their contents.
	writeTestJpeg(t, filepath.Join(src, "~home.jpg"), 240)
	require.NoError(t, os.WriteFile(filepath.Join(src, "~home.xmp"), []byte(testXmp), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(src, "~home.jpg.supplemental-metadata.json"), []byte("{}"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(src, "~home_edit.xmp"), []byte(testXmp), fs.ModeFile))
	writeTestJpeg(t, filepath.Join(src, "__IMG.JPG"), 244)
	require.NoError(t, os.WriteFile(filepath.Join(src, "__img.xmp"), []byte(testXmp), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(src, "~$x.docx"), []byte("x"), fs.ModeFile))
	writeTestJpeg(t, filepath.Join(src, "$keep.jpg"), 241)
	writeTestJpeg(t, filepath.Join(src, "._IMG_1234.JPG"), 242)
	writeTestJpeg(t, filepath.Join(src, "@eaDir", "x.jpg"), 243)
	require.NoError(t, os.WriteFile(filepath.Join(src, ".DS_Store"), []byte("x"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(src, "~$report.docx"), []byte("x"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(src, ".keep"), []byte{}, fs.ModeFile))

	// An empty folder whose name is not portable stays as well.
	require.NoError(t, os.MkdirAll(filepath.Join(src, "x:y"), fs.ModeDir))

	hook := captureLog(t)
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	imp.Start(ImportOptionsMove(src, ""))
	done := imp.Start(ImportOptionsMove(src, ""))

	for _, name := range []string{"a|b.jpg", "x:y.jpg", "caf\xe9.jpg", "a~b.jpg"} {
		assert.False(t, done[filepath.Join(src, name)].Exists(), "%q", name)
	}

	for _, name := range portable {
		assert.False(t, fs.FileExists(filepath.Join(src, name)), "%q was not imported", name)
	}

	for _, name := range skipped {
		assert.True(t, fs.FileExists(filepath.Join(src, name)), "%q was imported", name)
	}

	assert.DirExists(t, filepath.Join(src, "x:y"))
	assert.FileExists(t, filepath.Join(src, "~home.jpg"))
	assert.FileExists(t, filepath.Join(src, "~home.xmp"))
	assert.FileExists(t, filepath.Join(src, "~home.jpg.supplemental-metadata.json"))
	assert.FileExists(t, filepath.Join(src, "~home_edit.xmp"))
	assert.FileExists(t, filepath.Join(src, "__IMG.JPG"))
	assert.FileExists(t, filepath.Join(src, "__img.xmp"))
	assert.NoFileExists(t, filepath.Join(src, "~$x.docx"))
	assert.FileExists(t, filepath.Join(src, "$keep.jpg"))
	assert.FileExists(t, filepath.Join(src, "@eaDir", "x.jpg"))
	assert.FileExists(t, filepath.Join(src, ".keep"))
	assert.NoFileExists(t, filepath.Join(src, "._IMG_1234.JPG"))
	assert.NoFileExists(t, filepath.Join(src, ".DS_Store"))
	assert.NoFileExists(t, filepath.Join(src, "~$report.docx"))

	warnings := loggedMessages(hook, logrus.WarnLevel)

	for _, rel := range []string{"a|b.jpg", "x:y.jpg", "dir*", "caf\xe9.jpg", "a~b.jpg", "a$b.jpg", "pair.x|y.xmp", "stack.x|y.gif"} {
		assert.Equal(t, 1, countPrefix(warnings, "import: skipped "+clean.Log(rel)+" "), "%q in %v", rel, warnings)
	}

	assert.Zero(t, countPrefix(warnings, "import: skipped hidden file "+clean.Log("~home.xmp")))

	for _, rel := range []string{"~home.jpg", "$keep.jpg", "__IMG.JPG"} {
		assert.Equal(t, 1, countPrefix(warnings, "import: skipped hidden file "+clean.Log(rel)+", rename it"), rel)
	}

	for _, msg := range warnings {
		assert.NotContains(t, msg, "IMG_1234")
		assert.NotContains(t, msg, "report")
		assert.NotContains(t, msg, "eaDir")
	}
}

// countPrefix returns the number of messages that start with prefix.
func countPrefix(messages []string, prefix string) (n int) {
	for _, msg := range messages {
		if strings.HasPrefix(msg, prefix) {
			n++
		}
	}

	return n
}

// TestImport_StartUnportableLimit checks that names above the warning limit are counted in one warning per run.
func TestImport_StartUnportableLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	limitWarnedPaths(t, 1)

	cfg := config.TestConfig()
	src := t.TempDir()

	for i, name := range []string{"a|b.jpg", "c|d.jpg", "e$f.jpg"} {
		writeTestJpeg(t, filepath.Join(src, name), 260+i)
	}

	hook := captureLog(t)
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	imp.Start(ImportOptionsMove(src, ""))
	imp.Start(ImportOptionsMove(src, ""))

	warnings := loggedMessages(hook, logrus.WarnLevel)
	listed := countPrefixSuffix(warnings, "import: skipped", "rename it to import it")
	counted := countPrefix(warnings, "import: skipped 2 more names")

	assert.Equal(t, 1, listed)
	assert.Equal(t, 2, counted)

	// Names the import workers count for the index are reported by the same run.
	warnUnportable("index", entity.RootOriginals, "import-limit/x|y.jpg", "index it")
	hook.Reset()
	imp.Start(ImportOptionsMove(t.TempDir(), ""))

	assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.WarnLevel), "index: skipped 1 more name "))
}

// TestImport_StartUnportableFolder checks the names of folders below the import folder and of folders symbolic links point to.
func TestImport_StartUnportableFolder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	t.Run("Subfolder", func(t *testing.T) {
		// A folder below the import folder whose name is not portable is not imported, also if the import starts in it.
		dir := filepath.Join(cfg.ImportPath(), "unportable-"+rnd.Base36(8)+":x")
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		writeTestJpeg(t, filepath.Join(dir, "a.jpg"), 240)

		hook := captureLog(t)
		imp.Start(ImportOptionsMove(dir, ""))

		assert.True(t, fs.FileExists(filepath.Join(dir, "a.jpg")))
		assert.Contains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"),
			"import: skipped "+clean.Log(filepath.Base(dir))+" ")
	})
	t.Run("Links", func(t *testing.T) {
		src := t.TempDir()
		writeTestJpeg(t, filepath.Join(src, "real", "x.jpg"), 241)
		require.NoError(t, os.Symlink("real", filepath.Join(src, "A:link")))

		imp.Start(ImportOptionsMove(src, ""))

		assert.False(t, fs.FileExists(filepath.Join(src, "real", "x.jpg")), "real/x.jpg was not imported")
	})
	t.Run("LinkAfterFolder", func(t *testing.T) {
		// A folder whose name is not portable is imported under the portable name of a link that sorts after it.
		// The folder is resolved first, so a link in the path of the temporary folder does not hide a regression.
		src, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		writeTestJpeg(t, filepath.Join(src, "Best|Of", "y.jpg"), 252)
		require.NoError(t, os.Symlink("Best|Of", filepath.Join(src, "favorites")))
		hash := fs.Hash(filepath.Join(src, "Best|Of", "y.jpg"))
		t.Cleanup(func() { cleanupImportedHash(t, hash) })

		imp.Start(ImportOptionsMove(src, ""))

		assert.False(t, fs.FileExists(filepath.Join(src, "Best|Of", "y.jpg")), "favorites/y.jpg was not imported")
		_, err = entity.FirstFileByHash(hash)
		assert.NoError(t, err, "favorites/y.jpg is not indexed")
	})
}

// TestImport_DestinationFilenameUnportable checks that a file is not restored to a recorded name that is not portable.
func TestImport_DestinationFilenameUnportable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	// The file is indexed directly under a name that is not portable, then removed.
	fileName := filepath.Join(root, "old|name.jpg")
	writeTestJpeg(t, fileName, 250)
	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	require.NoError(t, ind.FileName(fileName, NewIndexOptions(dir, false, false, false, false, true, cfg)).Err)
	require.True(t, indexedFile(t, dir+"/old|name.jpg"))

	src := filepath.Join(t.TempDir(), "copy.jpg")
	writeTestJpeg(t, src, 250)
	require.NoError(t, os.Remove(fileName))

	mf, err := NewMediaFile(src)
	require.NoError(t, err)

	imp := NewImport(cfg, ind, convert)
	dest, err := imp.DestinationFilename(mf, mf, "")
	require.NoError(t, err)
	assert.NotEqual(t, fileName, dest)
	assert.True(t, clean.PortableName(filepath.Base(dest)), dest)
}

// TestImport_StartUnportableDest checks that an imported file is indexed under the name its destination pattern gives it.
func TestImport_StartUnportableDest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dest := cfg.Settings().Import.Dest
	t.Cleanup(func() { cfg.Settings().Import.Dest = dest })
	cfg.Settings().Import.Dest = "2006/01/2006-01-02T15:04:05_82F63B78.jpg"

	// The imported files are removed by the pattern of their names, also if a check below fails.
	t.Cleanup(func() {
		var files []entity.File
		_ = entity.UnscopedDb().Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, "%T%:%").Find(&files).Error
		for _, f := range files {
			_ = os.Remove(filepath.Join(cfg.OriginalsPath(), f.FileName))
			_ = entity.UnscopedDb().Delete(&entity.File{}, f.ID).Error
			_ = entity.UnscopedDb().Delete(&entity.Photo{}, f.PhotoID).Error
		}
	})

	src := t.TempDir()
	writeTestJpeg(t, filepath.Join(src, "colon.jpg"), 251)

	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	// An import run refuses the pattern and leaves the file where it is.
	imp.Start(ImportOptionsMove(src, ""))
	require.True(t, fs.FileExists(filepath.Join(src, "colon.jpg")), "colon.jpg was imported")

	// A job sent to the import worker directly, as sync does, still imports the file under the name the
	// pattern gives it.
	mf, err := NewMediaFile(filepath.Join(src, "colon.jpg"))
	require.NoError(t, err)
	related, err := mf.RelatedFiles(false)
	require.NoError(t, err)

	jobs := make(chan ImportJob)
	finished := make(chan struct{})

	go func() {
		ImportWorker(jobs)
		close(finished)
	}()

	jobs <- ImportJob{FileName: mf.FileName(), Related: related, IndexOpt: IndexOptionsAll(cfg),
		ImportOpt: ImportOptionsMove(src, ""), Imp: imp}
	close(jobs)
	<-finished

	require.False(t, fs.FileExists(filepath.Join(src, "colon.jpg")), "colon.jpg was not imported")

	var count int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, "%T%:%").Count(&count).Error)
	assert.Positive(t, count)
}

// TestImport_checkDest checks that destination folders and patterns giving names the index skips are refused.
func TestImport_CheckDest(t *testing.T) {
	cfg := config.TestConfig()
	imp := NewImport(cfg, nil, nil)
	saved := cfg.Settings().Import.Dest
	t.Cleanup(func() { cfg.Settings().Import.Dest = saved })

	t.Run("Supported", func(t *testing.T) {
		cfg.Settings().Import.Dest = ""
		for _, folder := range []string{"", ".", "./x", "x/./y", "/x/", "2019", "users/jane.doe", "import/sub"} {
			assert.NoError(t, imp.CheckDest(folder), folder)
		}
	})
	t.Run("Folder", func(t *testing.T) {
		cfg.Settings().Import.Dest = ""
		for _, folder := range []string{"a:b", "users/50%off", "users/__bob", "@eaDir/x", "~home", "..", "../x", "x/../../y"} {
			err := imp.CheckDest(folder)
			assert.ErrorIs(t, err, ErrImportDest, folder)
			assert.ErrorContains(t, err, "folder", folder)
		}
	})
	t.Run("Message", func(t *testing.T) {
		cfg.Settings().Import.Dest = ""
		assert.EqualError(t, imp.CheckDest("a:b"), "destination folder a:b is not supported")
		cfg.Settings().Import.Dest = "2006/~01/20060102_150405_82F63B78.jpg"
		assert.EqualError(t, imp.CheckDest(""), "destination pattern 2006/~01/20060102_150405_82F63B78.jpg is not supported")
	})
	t.Run("Pattern", func(t *testing.T) {
		cfg.Settings().Import.Dest = "2006/01/2006-01-02T15:04:05_82F63B78.jpg"
		assert.ErrorIs(t, imp.CheckDest(""), ErrImportDest)
		cfg.Settings().Import.Dest = "2006/~01/20060102_150405_82F63B78.jpg"
		assert.ErrorIs(t, imp.CheckDest(""), ErrImportDest)
		cfg.Settings().Import.Dest = "2006/01/02/20060102-150405-82F63B78.jpg"
		assert.NoError(t, imp.CheckDest(""))
	})
}

// TestImport_StartUnsupportedDest checks that an import with an unsupported destination leaves the files in place.
func TestImport_StartUnsupportedDest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	src := t.TempDir()
	writeTestJpeg(t, filepath.Join(src, "ok.jpg"), 270)

	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	// Callers get the error without the details, which name folders.
	_, err := imp.Run(ImportOptionsMove(src, "a:b"))
	assert.Equal(t, ErrImportDest, err)
	assert.FileExists(t, filepath.Join(src, "ok.jpg"))
	assert.NoDirExists(t, filepath.Join(cfg.OriginalsPath(), "a:b"))
}
