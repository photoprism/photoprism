package photoprism

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestWarnedPaths_First(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		w := &warnedPaths{limit: 10}
		assert.True(t, w.First("index", "originals:a|b"))
		assert.False(t, w.First("index", "originals:a|b"))
		assert.True(t, w.First("import", "import:a|b"))

		// The same path is remembered separately for each prefix.
		assert.True(t, w.First("import hidden", "originals:a|b"))
		assert.False(t, w.First("import hidden", "originals:a|b"))
	})
	t.Run("Limit", func(t *testing.T) {
		// Paths above the limit are not remembered, and are counted for their prefix instead.
		w := &warnedPaths{limit: 3}
		for i := 0; i < 3; i++ {
			assert.True(t, w.First("index", strconv.Itoa(i)))
		}
		assert.False(t, w.First("index", "3"))
		assert.False(t, w.First("index", "3"))
		assert.False(t, w.First("import", "4"))
		assert.False(t, w.First("index", "0"))
		assert.Len(t, w.seen, 3)
		assert.Equal(t, 1, w.Flush("index"))
		assert.Equal(t, 0, w.Flush("index"))
		assert.Equal(t, 1, w.Flush("import"))

		// A path counted before is counted again by the next run, and never reported on its own.
		assert.False(t, w.First("index", "3"))
		assert.Equal(t, 1, w.Flush("index"))
	})
	t.Run("CountLimit", func(t *testing.T) {
		// Counted paths are remembered up to the limit, and counted without it above.
		w := &warnedPaths{limit: 1}
		assert.True(t, w.First("index", "0"))
		assert.False(t, w.First("index", "1"))
		assert.False(t, w.First("index", "1"))
		assert.False(t, w.First("index", "2"))
		assert.False(t, w.First("index", "2"))
		assert.Len(t, w.more["index"], 1)
		assert.Equal(t, 3, w.Flush("index"))
		assert.Equal(t, 0, w.Flush("index"))
	})
}

func TestFlushUnportable(t *testing.T) {
	limitWarnedPaths(t, 1)

	hook := captureLog(t)

	for run := 0; run < 2; run++ {
		for _, rel := range []string{"a|b", "c|d", "e|f"} {
			warnUnportable("index", "originals", rel, "index it")
		}
		flushUnportable("index")
	}

	flushUnportable("index")

	warnings := loggedMessages(hook, logrus.WarnLevel)
	assert.Equal(t, 1, countPrefixSuffix(warnings, "index: skipped", "rename it to index it"))
	assert.Equal(t, 2, countPrefix(warnings, "index: skipped 2 more names with unsupported characters"))
}

func TestWarnHidden(t *testing.T) {
	limitWarnedPaths(t, 1)

	hook := captureLog(t)

	for run := 0; run < 2; run++ {
		for _, rel := range []string{"~a.jpg", "$b.jpg", "@c.jpg"} {
			warnHidden("import", "import", rel, "import it")
		}
		flushHidden("import")
	}

	// Hidden files are counted apart from names that are not portable.
	flushUnportable("import")

	warnings := loggedMessages(hook, logrus.WarnLevel)
	assert.Equal(t, []string{
		"import: skipped hidden file ~a.jpg, rename it to import it",
		"import: skipped 2 more hidden files",
		"import: skipped 2 more hidden files",
	}, warnings)
}

func TestFlushHidden(t *testing.T) {
	limitWarnedPaths(t, 0)

	hook := captureLog(t)
	flushHidden("import")
	warnHidden("import", "import", "~a.jpg", "import it")
	flushHidden("import")
	flushHidden("import")

	assert.Equal(t, []string{"import: skipped 1 more hidden file"}, loggedMessages(hook, logrus.WarnLevel))
}

func TestFileIndexed(t *testing.T) {
	dir := filepath.Join(Config().OriginalsPath(), "file-indexed-"+rnd.Base36(8))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	writeTestJpeg(t, filepath.Join(dir, "a.jpg"), 300)
	mf, err := NewMediaFile(filepath.Join(dir, "a.jpg"))
	require.NoError(t, err)

	assert.False(t, fileIndexed(mf))

	// Rows without a root match as well, as in the index.
	row := &entity.File{FileUID: rnd.GenerateUID(entity.FileUID), FileRoot: "", FileName: mf.RootRelName(), FileHash: rnd.GenerateUID('h')}
	require.NoError(t, entity.UnscopedDb().Create(row).Error)
	t.Cleanup(func() { assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(row).Error) })
	require.NoError(t, entity.UnscopedDb().Model(row).UpdateColumn("file_root", "").Error)
	assert.True(t, fileIndexed(mf))

	require.NoError(t, entity.UnscopedDb().Delete(row).Error)
	assert.False(t, fileIndexed(mf))
}

func TestSkipUnportableDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, fs.MkdirAll(filepath.Join(dir, "a|b")))
	require.NoError(t, os.Symlink(filepath.Join(dir, "a|b"), filepath.Join(dir, "x:y")))

	t.Run("Portable", func(t *testing.T) {
		done := fs.Done{filepath.Join(dir, "ok"): fs.Found}
		assert.Equal(t, "", skipUnportableDir(done, filepath.Join(dir, "ok"), "ok", false))
		assert.Len(t, done, 1)
	})
	t.Run("Folder", func(t *testing.T) {
		done := fs.Done{filepath.Join(dir, "a|b"): fs.Found, "other": fs.Found}
		assert.Equal(t, "a|b", skipUnportableDir(done, filepath.Join(dir, "a|b"), "a|b", false))
		assert.Equal(t, fs.Done{"other": fs.Found}, done)
	})
	t.Run("Link", func(t *testing.T) {
		resolved, err := filepath.EvalSymlinks(filepath.Join(dir, "a|b"))
		require.NoError(t, err)
		done := fs.Done{filepath.Join(dir, "x:y"): fs.Found, resolved: fs.Found}
		assert.Equal(t, "sub/x:y", skipUnportableDir(done, filepath.Join(dir, "x:y"), "sub/x:y/z", true))
		assert.Empty(t, done)
	})
}

func TestWarnUnportable(t *testing.T) {
	limitWarnedPaths(t, 10)
	hook := captureLog(t)

	warnUnportable("index", "originals", "warn-unportable/a|x", "index it")
	warnUnportable("index", "originals", "warn-unportable/a|x", "index it")

	assert.Equal(t, []string{"index: skipped warn-unportable/a|x because of unsupported characters, rename it to index it"},
		loggedMessages(hook, logrus.WarnLevel))
}

// limitWarnedPaths replaces the remembered paths with an empty set that holds up to limit paths during a test.
func limitWarnedPaths(t *testing.T, limit int) {
	saved := unsupportedPaths
	t.Cleanup(func() { unsupportedPaths = saved })
	unsupportedPaths = &warnedPaths{limit: limit}
}

// countPrefixSuffix returns the number of messages that start with prefix and end with suffix.
func countPrefixSuffix(messages []string, prefix, suffix string) (n int) {
	for _, msg := range messages {
		if strings.HasPrefix(msg, prefix) && strings.HasSuffix(msg, suffix) {
			n++
		}
	}

	return n
}

func TestPortableRelated(t *testing.T) {
	dir := t.TempDir()
	newFile := func(name string) *MediaFile {
		fileName := filepath.Join(dir, name)
		writeTestJpeg(t, fileName, len(name))
		mf, err := NewMediaFile(fileName)
		require.NoError(t, err)
		return mf
	}

	main, sidecar, other := newFile("a.jpg"), newFile("a.x|y.jpg"), newFile("a.b.jpg")
	var skipped []string
	report := func(f *MediaFile) { skipped = append(skipped, f.BaseName()) }

	t.Run("Files", func(t *testing.T) {
		skipped = nil
		related := PortableRelated(RelatedFiles{Main: main, Files: MediaFiles{main, sidecar, other}}, main, report)
		assert.Equal(t, main, related.Main)
		assert.Equal(t, MediaFiles{main, other}, related.Files)
		assert.Equal(t, []string{"a.x|y.jpg"}, skipped)
	})
	t.Run("Main", func(t *testing.T) {
		skipped = nil
		related := PortableRelated(RelatedFiles{Main: sidecar, Files: MediaFiles{main, sidecar}}, main, report)
		assert.Equal(t, main, related.Main)
		assert.Equal(t, MediaFiles{main}, related.Files)
		assert.Equal(t, []string{"a.x|y.jpg"}, skipped)
	})
	t.Run("NotInGroup", func(t *testing.T) {
		// A file the group leaves out, e.g. in a format that is turned off, does not become the main file.
		related := PortableRelated(RelatedFiles{Main: sidecar, Files: MediaFiles{other, sidecar}}, main, report)
		assert.Nil(t, related.Main)
		assert.Equal(t, MediaFiles{other}, related.Files)
	})
	t.Run("NoMain", func(t *testing.T) {
		related := PortableRelated(RelatedFiles{Main: sidecar, Files: MediaFiles{sidecar}}, sidecar, report)
		assert.Nil(t, related.Main)
		assert.Empty(t, related.Files)
	})
}

func TestPortableSiblings(t *testing.T) {
	dir := t.TempDir()
	newFile := func(name string) *MediaFile {
		fileName := filepath.Join(dir, name)
		writeTestJpeg(t, fileName, len(name)+100)
		mf, err := NewMediaFile(fileName)
		require.NoError(t, err)
		return mf
	}

	main, sibling, other, legacy := newFile("s.jpg"), newFile("s.x|y.jpg"), newFile("s.b.jpg"), newFile("c$d.jpg")
	var skipped []string
	report := func(f *MediaFile) { skipped = append(skipped, f.BaseName()) }

	t.Run("Siblings", func(t *testing.T) {
		skipped = nil
		related := PortableSiblings(RelatedFiles{Main: main, Files: MediaFiles{main, sibling, other}}, main, report)
		assert.Equal(t, main, related.Main)
		assert.Equal(t, MediaFiles{main, other}, related.Files)
		assert.Equal(t, []string{"s.x|y.jpg"}, skipped)
	})
	t.Run("KeepsFile", func(t *testing.T) {
		skipped = nil
		related := PortableSiblings(RelatedFiles{Main: legacy, Files: MediaFiles{legacy}}, legacy, report)
		assert.Equal(t, legacy, related.Main)
		assert.Equal(t, MediaFiles{legacy}, related.Files)
		assert.Empty(t, skipped)
	})
	t.Run("MainSibling", func(t *testing.T) {
		skipped = nil
		related := PortableSiblings(RelatedFiles{Main: sibling, Files: MediaFiles{main, sibling}}, main, report)
		assert.Equal(t, main, related.Main)
		assert.Equal(t, MediaFiles{main}, related.Files)
	})
	t.Run("NotInGroup", func(t *testing.T) {
		skipped = nil
		related := PortableSiblings(RelatedFiles{Main: sibling, Files: MediaFiles{other, sibling}}, main, report)
		assert.Nil(t, related.Main)
		assert.Equal(t, MediaFiles{other}, related.Files)
	})
	t.Run("Folder", func(t *testing.T) {
		// Other files in originals must be in folders whose names are portable, generated files need not be.
		folder := "portable-siblings-" + strconv.Itoa(len(dir)) + "/dir$"
		originals := filepath.Join(Config().OriginalsPath(), folder)
		sidecar := filepath.Join(Config().SidecarPath(), folder)
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Dir(originals))
			_ = os.RemoveAll(filepath.Dir(sidecar))
		})

		newIn := func(dir, name string) *MediaFile {
			fileName := filepath.Join(dir, name)
			writeTestJpeg(t, fileName, len(name)+200)
			mf, err := NewMediaFile(fileName)
			require.NoError(t, err)
			return mf
		}

		file, other, preview := newIn(originals, "k.jpg"), newIn(originals, "k.b.jpg"), newIn(sidecar, "k.cr2.jpg")

		skipped = nil
		related := PortableSiblings(RelatedFiles{Main: file, Files: MediaFiles{file, other, preview}}, file, report)
		assert.Equal(t, file, related.Main)
		assert.Equal(t, MediaFiles{file, preview}, related.Files)
		assert.Equal(t, []string{"k.b.jpg"}, skipped)

		// The index walk checks folders on its own, so PortableRelated does not.
		related = PortableRelated(RelatedFiles{Main: other, Files: MediaFiles{other}}, other, report)
		assert.Equal(t, other, related.Main)

		// A file in the sidecar folder does not replace a main file in a folder whose name is not portable, which
		// stays if it is indexed. A file in originals does replace one that is not indexed.
		raw := newIn(originals, "k.cr2")

		skipped = nil
		related = PortableSiblings(RelatedFiles{Main: raw, Files: MediaFiles{raw, preview}}, preview, report)
		assert.Nil(t, related.Main)
		assert.Equal(t, MediaFiles{preview}, related.Files)
		assert.Equal(t, []string{"k.cr2"}, skipped)

		related = PortableSiblings(RelatedFiles{Main: raw, Files: MediaFiles{raw, file}}, file, report)
		assert.Equal(t, file, related.Main)
		assert.Equal(t, MediaFiles{file}, related.Files)

		row := &entity.File{FileUID: rnd.GenerateUID(entity.FileUID), FileRoot: entity.RootOriginals, FileName: folder + "/k.cr2",
			FileHash: rnd.GenerateUID('h')}
		require.NoError(t, entity.UnscopedDb().Create(row).Error)
		t.Cleanup(func() { assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(row).Error) })

		skipped = nil
		related = PortableSiblings(RelatedFiles{Main: raw, Files: MediaFiles{raw, preview}}, preview, report)
		assert.Equal(t, raw, related.Main)
		assert.Equal(t, MediaFiles{raw, preview}, related.Files)
		assert.Empty(t, skipped)

		related = PortableSiblings(RelatedFiles{Main: raw, Files: MediaFiles{raw, file}}, file, report)
		assert.Equal(t, raw, related.Main)
	})
}
