package photoprism

import (
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
		assert.True(t, w.First("originals:a|b"))
		assert.False(t, w.First("originals:a|b"))
		assert.True(t, w.First("import:a|b"))
	})
	t.Run("Limit", func(t *testing.T) {
		w := &warnedPaths{limit: 3}
		for i := 0; i < 4; i++ {
			assert.True(t, w.First(strconv.Itoa(i)))
		}
		assert.Len(t, w.seen, 3)
		assert.True(t, w.First("0"))
		assert.False(t, w.First("3"))
	})
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
	rel := "warn-unportable/" + strconv.Itoa(len(unportablePaths.order)) + "|x"

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
		related := portableRelated(RelatedFiles{Main: main, Files: MediaFiles{main, sidecar, other}}, main, report)
		assert.Equal(t, main, related.Main)
		assert.Equal(t, MediaFiles{main, other}, related.Files)
		assert.Equal(t, []string{"a.x|y.jpg"}, skipped)
	})
	t.Run("Main", func(t *testing.T) {
		skipped = nil
		related := portableRelated(RelatedFiles{Main: sidecar, Files: MediaFiles{main, sidecar}}, main, report)
		assert.Equal(t, main, related.Main)
		assert.Equal(t, MediaFiles{main}, related.Files)
		assert.Equal(t, []string{"a.x|y.jpg"}, skipped)
	})
	t.Run("NotInGroup", func(t *testing.T) {
		// A file the group leaves out, e.g. in a format that is turned off, does not become the main file.
		related := portableRelated(RelatedFiles{Main: sidecar, Files: MediaFiles{other, sidecar}}, main, report)
		assert.Nil(t, related.Main)
		assert.Equal(t, MediaFiles{other}, related.Files)
	})
	t.Run("NoMain", func(t *testing.T) {
		related := portableRelated(RelatedFiles{Main: sidecar, Files: MediaFiles{sidecar}}, sidecar, report)
		assert.Nil(t, related.Main)
		assert.Empty(t, related.Files)
	})
}
