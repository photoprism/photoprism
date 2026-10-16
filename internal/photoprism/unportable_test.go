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
)

func TestWarnedPaths_First(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		w := &warnedPaths{limit: 10}
		assert.True(t, w.First("index", "originals:a|b"))
		assert.False(t, w.First("index", "originals:a|b"))
		assert.True(t, w.First("import", "import:a|b"))
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
	saved := unportablePaths
	t.Cleanup(func() { unportablePaths = saved })
	unportablePaths = &warnedPaths{limit: 1}

	hook := captureLog(t)

	// count returns the number of warnings that start with prefix and contain text.
	count := func(prefix, text string) (n int) {
		for _, msg := range loggedMessages(hook, logrus.WarnLevel) {
			if strings.HasPrefix(msg, prefix) && strings.Contains(msg, text) {
				n++
			}
		}
		return n
	}

	for run := 0; run < 2; run++ {
		for _, rel := range []string{"a|b", "c|d", "e|f"} {
			warnUnportable("index", "originals", rel, "index it")
		}
		flushUnportable("index")
	}

	flushUnportable("index")

	assert.Equal(t, 1, count("index: skipped", "rename it to index it"))
	assert.Equal(t, 2, count("index: skipped 2 more names", "characters that are not supported"))
}

func TestWarnHidden(t *testing.T) {
	saved := unportablePaths
	t.Cleanup(func() { unportablePaths = saved })
	unportablePaths = &warnedPaths{limit: 1}

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
	saved := unportablePaths
	t.Cleanup(func() { unportablePaths = saved })
	unportablePaths = &warnedPaths{limit: 0}

	hook := captureLog(t)
	flushHidden("import")
	warnHidden("import", "import", "~a.jpg", "import it")
	flushHidden("import")
	flushHidden("import")

	assert.Equal(t, []string{"import: skipped 1 more hidden file"}, loggedMessages(hook, logrus.WarnLevel))
}

func TestUnportablePath(t *testing.T) {
	t.Run("Portable", func(t *testing.T) {
		for _, rel := range []string{"", "a", "2020/03 - Março/#1 & Friends.jpg", "a/b/c"} {
			assert.Equal(t, "", unportablePath(rel), rel)
		}
	})
	t.Run("NotPortable", func(t *testing.T) {
		assert.Equal(t, "a|b", unportablePath("a|b"))
		assert.Equal(t, "2020/dir*", unportablePath("2020/dir*/inner.jpg"))
		assert.Equal(t, "2020/x:y.jpg", unportablePath("2020/x:y.jpg"))
		assert.Equal(t, "~home", unportablePath("~home/a/b"))
		assert.Equal(t, "a/caf\xe9", unportablePath("a/caf\xe9/b"))
	})
}

func TestWarnUnportable(t *testing.T) {
	hook := captureLog(t)
	rel := "warn-unportable/" + strconv.Itoa(len(unportablePaths.seen)) + "|x"

	warnUnportable("index", "originals", rel, "index it")
	warnUnportable("index", "originals", rel, "index it")

	count := 0

	for _, msg := range loggedMessages(hook, logrus.WarnLevel) {
		if strings.HasPrefix(msg, "index: skipped") && strings.Contains(msg, rel) && strings.HasSuffix(msg, "rename it to index it") {
			count++
		}
	}

	assert.Equal(t, 1, count)
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
	})
}

func TestUnsupportedPath(t *testing.T) {
	t.Run("Supported", func(t *testing.T) {
		for _, rel := range []string{"", "2019/10", "users/jane.doe", "a_b/c"} {
			assert.Equal(t, "", unsupportedPath(rel), rel)
		}
	})
	t.Run("NotSupported", func(t *testing.T) {
		assert.Equal(t, "2019/a:b", unsupportedPath("2019/a:b/c"))
		assert.Equal(t, "users/__bob", unsupportedPath("users/__bob/x"))
		assert.Equal(t, "@eaDir", unsupportedPath("@eaDir"))
		assert.Equal(t, "a/$x", unsupportedPath("a/$x/b"))
	})
}
