package photoprism

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// writeTestJpeg writes a small JPEG whose pixels depend on seed, so that no two test files are duplicates.
func writeTestJpeg(t *testing.T, fileName string, seed int) {
	t.Helper()

	writeTestImage(t, fileName, seed, func(w io.Writer, img image.Image) error { return jpeg.Encode(w, img, nil) })
}

// writeTestPng writes a small PNG whose pixels depend on seed, like writeTestJpeg.
func writeTestPng(t *testing.T, fileName string, seed int) {
	t.Helper()

	writeTestImage(t, fileName, seed, png.Encode)
}

// writeTestImage writes a small image whose pixels depend on seed with the given encoder.
func writeTestImage(t *testing.T, fileName string, seed int, encode func(io.Writer, image.Image) error) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(fileName), fs.ModeDir))

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	shade := byte(seed % 256) //nolint:gosec // the remainder fits a byte.

	for x := 0; x < 32; x++ {
		for y := 0; y < 32; y++ {
			img.Set(x, y, color.RGBA{R: shade * 37, G: byte(x) * 8, B: byte(y)*8 + shade, A: 255}) //nolint:gosec // x and y are below 32.
		}
	}

	f, err := os.Create(fileName) //nolint:gosec // fileName is below the test originals folder.
	require.NoError(t, err)
	require.NoError(t, encode(f, img))
	require.NoError(t, f.Close())
}

// writeTestGif writes a small GIF image.
func writeTestGif(t *testing.T, fileName string) {
	t.Helper()

	img := image.NewPaletted(image.Rect(0, 0, 16, 16), color.Palette{color.Black, color.White})
	f, err := os.Create(fileName) //nolint:gosec // fileName is below the test originals folder.
	require.NoError(t, err)
	require.NoError(t, gif.Encode(f, img, nil))
	require.NoError(t, f.Close())
}

// testXmp is a minimal XMP sidecar.
const testXmp = `<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>Pair</dc:title></rdf:Description></rdf:RDF></x:xmpmeta>`

// cleanupIndexed removes the index entries below a folder of the originals, so later tests do not see them.
func cleanupIndexed(t *testing.T, dir string) {
	t.Cleanup(func() {
		var photoIDs []uint
		db := entity.UnscopedDb()
		_ = db.Model(&entity.File{}).Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, dir+"/%").Pluck("photo_id", &photoIDs).Error
		_ = db.Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, dir+"/%").Delete(&entity.File{}).Error
		if len(photoIDs) > 0 {
			_ = db.Where("id IN (?)", photoIDs).Delete(&entity.Photo{}).Error
		}
		_ = db.Where("root = ? AND (path = ? OR path LIKE ?)", entity.RootOriginals, dir, dir+"/%").Delete(&entity.Folder{}).Error
	})
}

// indexedFile reports whether a file below the originals folder has an index entry.
func indexedFile(t *testing.T, rel string) bool {
	t.Helper()

	var count int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).
		Where("file_root = ? AND file_name = ?", entity.RootOriginals, rel).Count(&count).Error)

	return count > 0
}

// TestIndex_StartUnportableNames checks that the index skips files and folders whose names are not portable.
func TestIndex_StartUnportableNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	portable := []string{"ok.jpg", "03 - Março/#1 & Friends [2020] (copy) Tom's 😀.jpg", "日本\u3000旅行/Screenshot 10.00.00\u202fAM.jpg"}
	skipped := []string{"a|b.jpg", "x:y.jpg", "dir*/inner.jpg", "caf\xe9.jpg", "a~b.jpg", "price$5.jpg"}

	for i, name := range append(append([]string{}, portable...), skipped...) {
		writeTestJpeg(t, filepath.Join(root, name), i+1)
	}

	// A sidecar whose name is not portable is not indexed with its picture.
	writeTestJpeg(t, filepath.Join(root, "pair.jpg"), 50)
	require.NoError(t, os.WriteFile(filepath.Join(root, "pair.x|y.xmp"), []byte(testXmp), fs.ModeFile))
	portable = append(portable, "pair.jpg")
	skipped = append(skipped, "pair.x|y.xmp")

	// A picture is indexed if the file that would be the main file of its group has a name that is not portable.
	writeTestJpeg(t, filepath.Join(root, "stack.jpg"), 60)
	writeTestGif(t, filepath.Join(root, "stack.x|y.gif"))
	portable = append(portable, "stack.jpg")
	skipped = append(skipped, "stack.x|y.gif")

	// Hidden and ignored names keep their handling and are not reported.
	writeTestJpeg(t, filepath.Join(root, ".x|y", "hidden.jpg"), 70)
	writeTestJpeg(t, filepath.Join(root, "$RECYCLE.BIN", "deleted.jpg"), 72)
	writeTestJpeg(t, filepath.Join(root, "~$draft.jpg"), 73)
	writeTestJpeg(t, filepath.Join(root, "skip|me.jpg"), 71)
	require.NoError(t, os.WriteFile(filepath.Join(root, fs.PPIgnoreFilename), []byte("skip|me.jpg\n"), fs.ModeFile))

	hook := captureLog(t)
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := NewIndexOptions(dir, false, false, false, false, true, cfg)
	opt.FaceRegeneration = &FaceRegeneration{}

	ind.Start(opt)
	found, _ := ind.Start(opt)

	// Skipped files are reported to a face regeneration as files excluded from indexing.
	for _, name := range []string{"a|b.jpg", "x:y.jpg", "dir*/inner.jpg", "pair.x|y.xmp", "stack.x|y.gif"} {
		assert.True(t, opt.FaceRegeneration.Skipped(filepath.Join(root, name)), "%q", name)
	}

	assert.False(t, opt.FaceRegeneration.Skipped(filepath.Join(root, "ok.jpg")))

	for _, name := range []string{"a|b.jpg", "x:y.jpg", "caf\xe9.jpg", "a~b.jpg"} {
		assert.False(t, found[filepath.Join(root, name)].Exists(), "%q", name)
	}

	// Hidden files are recorded as found, so a purge does not flag earlier index entries as missing.
	assert.True(t, found[filepath.Join(root, "~$draft.jpg")].Exists())
	assert.False(t, indexedFile(t, dir+"/~$draft.jpg"))
	assert.False(t, indexedFile(t, dir+"/$RECYCLE.BIN/deleted.jpg"))

	for _, name := range portable {
		assert.True(t, indexedFile(t, dir+"/"+name), name)
	}

	for _, name := range skipped {
		// PostgreSQL cannot compare a name that is not valid UTF-8, which the walk did not reach either.
		if utf8.ValidString(name) {
			assert.False(t, indexedFile(t, dir+"/"+name), "%q", name)
		}
	}

	// Portable folders get their rows.
	for _, name := range []string{"03 - Março", "日本\u3000旅行"} {
		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Folder{}).
			Where("root = ? AND path = ?", entity.RootOriginals, dir+"/"+name).Count(&count).Error)
		assert.Equal(t, int64(1), count, name)
	}

	// Each skipped path is reported once, a folder for its whole contents.
	warnings := loggedMessages(hook, logrus.WarnLevel)

	for _, rel := range []string{dir + "/a|b.jpg", dir + "/x:y.jpg", dir + "/dir*", dir + "/caf\xe9.jpg", dir + "/a~b.jpg", dir + "/price$5.jpg", dir + "/pair.x|y.xmp", dir + "/stack.x|y.gif"} {
		assert.Equal(t, 1, countPrefix(warnings, "index: skipped "+clean.Log(rel)+" "), "%q in %v", rel, warnings)
	}

	for _, msg := range warnings {
		assert.NotContains(t, msg, "skipped "+clean.Log(dir+"/.x|y")+" ")
		assert.NotContains(t, msg, "skip|me")
		assert.NotContains(t, msg, "RECYCLE")
		assert.NotContains(t, msg, "draft")
	}

	var folders int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.Folder{}).
		Where("root = ? AND path LIKE ?", entity.RootOriginals, dir+"/dir%").Count(&folders).Error)
	assert.Zero(t, folders)
}

// TestIndex_StartUnportableIndexed checks that a file indexed under a name that is not portable stays in the library.
func TestIndex_StartUnportableIndexed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })

	cleanupIndexed(t, dir)

	rel := dir + "/old|name.jpg"
	fileName := filepath.Join(cfg.OriginalsPath(), rel)
	writeTestJpeg(t, fileName, 99)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := NewIndexOptions(dir, false, false, false, false, true, cfg)

	// The file is indexed directly, which gives it an index entry the walk does not create.
	result := ind.FileName(fileName, opt)
	require.NoError(t, result.Err)
	require.True(t, indexedFile(t, rel))

	var before entity.File
	require.NoError(t, entity.UnscopedDb().Where("file_root = ? AND file_name = ?", entity.RootOriginals, rel).First(&before).Error)

	// A changed file is not updated.
	writeTestJpeg(t, fileName, 98)
	modified := time.Now().Add(time.Hour)
	require.NoError(t, os.Chtimes(fileName, modified, modified))

	hook := captureLog(t)
	found, _ := ind.Start(opt)

	purge := NewPurge(cfg, NewFiles())
	_, _, _, err := purge.Start(PurgeOptions{Path: dir, Ignore: found})
	require.NoError(t, err)

	var after entity.File
	require.NoError(t, entity.UnscopedDb().Where("file_root = ? AND file_name = ?", entity.RootOriginals, rel).First(&after).Error)
	assert.False(t, after.FileMissing)
	assert.False(t, after.DeletedAt.Valid)
	assert.Equal(t, before.PhotoID, after.PhotoID)
	assert.Equal(t, before.FileHash, after.FileHash)

	photo := entity.FindPhoto(entity.Photo{ID: after.PhotoID})
	require.NotNil(t, photo)
	assert.False(t, photo.DeletedAt.Valid)

	assert.Contains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "index: skipped "+rel)
}

// TestIndex_StartUnportableLinks checks that a folder a symbolic link points to is indexed under the name that is portable.
func TestIndex_StartUnportableLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	// The link with the name that is not portable sorts before the folder it points to, and the other way round.
	writeTestJpeg(t, filepath.Join(root, "real", "x.jpg"), 80)
	require.NoError(t, os.Symlink("real", filepath.Join(root, "A:link")))
	writeTestJpeg(t, filepath.Join(root, "Best|Of", "y.jpg"), 81)
	require.NoError(t, os.Symlink("Best|Of", filepath.Join(root, "favorites")))

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	ind.Start(NewIndexOptions(dir, false, false, false, false, true, cfg))

	assert.True(t, indexedFile(t, dir+"/real/x.jpg"))
	assert.False(t, indexedFile(t, dir+"/A:link/x.jpg"))
	assert.True(t, indexedFile(t, dir+"/favorites/y.jpg"))
	assert.False(t, indexedFile(t, dir+"/Best|Of/y.jpg"))
}

// TestIndex_StartUnportablePath checks that a folder given to the index whose name is not portable is skipped.
func TestIndex_StartUnportablePath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	writeTestJpeg(t, filepath.Join(root, "dir*", "a.jpg"), 90)
	writeTestJpeg(t, filepath.Join(root, "dir*", "sub", "b.jpg"), 91)

	hook := captureLog(t)
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	ind.Start(NewIndexOptions(dir+"/dir*", false, false, false, false, true, cfg))
	ind.Start(NewIndexOptions(dir+"/dir*/sub", false, false, false, false, true, cfg))

	assert.False(t, indexedFile(t, dir+"/dir*/a.jpg"))
	assert.False(t, indexedFile(t, dir+"/dir*/sub/b.jpg"))

	assert.Equal(t, 1, countPrefix(loggedMessages(hook, logrus.WarnLevel), "index: skipped "+clean.Log(dir+"/dir*")+" "))
}

// TestIndex_StartUnportableLimit checks that names above the warning limit are counted in one warning per run.
func TestIndex_StartUnportableLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	limitWarnedPaths(t, 1)

	cfg := config.TestConfig()
	dir := "unportable-limit-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	for i, name := range []string{"a|b.jpg", "c|d.jpg", "e$f.jpg"} {
		writeTestJpeg(t, filepath.Join(root, name), i+80)
	}

	hook := captureLog(t)
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := NewIndexOptions(dir, false, false, false, false, true, cfg)

	ind.Start(opt)
	ind.Start(opt)

	warnings := loggedMessages(hook, logrus.WarnLevel)
	listed := countPrefixSuffix(warnings, "index: skipped", "rename it to index it")
	counted := countPrefix(warnings, "index: skipped 2 more names")

	assert.Equal(t, 1, listed)
	assert.Equal(t, 2, counted)
}

// TestIndex_FileNameUnportable checks that reindexing a single file leaves related files whose names are not
// portable out, but not the file itself.
func TestIndex_FileNameUnportable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	dir := "unportable-single-" + rnd.Base36(8)
	root := filepath.Join(cfg.OriginalsPath(), dir)
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	cleanupIndexed(t, dir)

	writeTestJpeg(t, filepath.Join(root, "IMG.jpg"), 90)
	writeTestGif(t, filepath.Join(root, "IMG.x|y.gif"))
	writeTestJpeg(t, filepath.Join(root, "c$d.jpg"), 91)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	result := ind.FileName(filepath.Join(root, "IMG.jpg"), IndexOptionsSingle(cfg))
	require.NoError(t, result.Err)
	assert.True(t, indexedFile(t, dir+"/IMG.jpg"))
	assert.False(t, indexedFile(t, dir+"/IMG.x|y.gif"))

	// A file that is reindexed on its own is indexed whatever its name.
	result = ind.FileName(filepath.Join(root, "c$d.jpg"), IndexOptionsSingle(cfg))
	require.NoError(t, result.Err)
	assert.True(t, indexedFile(t, dir+"/c$d.jpg"))

	// So is a file in a folder whose name is not portable, but not the other files in that folder.
	writeTestJpeg(t, filepath.Join(root, "f$g", "L.jpg"), 92)
	writeTestJpeg(t, filepath.Join(root, "f$g", "L.b.jpg"), 93)

	result = ind.FileName(filepath.Join(root, "f$g", "L.jpg"), IndexOptionsSingle(cfg))
	require.NoError(t, result.Err)
	assert.True(t, indexedFile(t, dir+"/f$g/L.jpg"))
	assert.False(t, indexedFile(t, dir+"/f$g/L.b.jpg"))

	// Reindexing fails if no file is left that can be the main file.
	disableRaw := cfg.Options().DisableRaw
	t.Cleanup(func() { cfg.Options().DisableRaw = disableRaw })
	cfg.Options().DisableRaw = true

	require.NoError(t, os.WriteFile(filepath.Join(root, "R.cr2"), []byte("raw"), fs.ModeFile))
	writeTestJpeg(t, filepath.Join(root, "R.x|y.jpg"), 94)

	result = ind.FileName(filepath.Join(root, "R.cr2"), IndexOptionsSingle(cfg))
	assert.Error(t, result.Err)
	assert.Equal(t, IndexFailed, result.Status)
	assert.False(t, indexedFile(t, dir+"/R.x|y.jpg"))
}
