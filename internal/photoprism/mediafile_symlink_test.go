package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// annexLibrary builds the shape a git-annex library has: every original is a symbolic link into the
// annex object store, and a file whose content has been dropped is a link with nothing behind it.
// Returns the library root and the object store path.
func annexLibrary(t *testing.T) (root, objects string) {
	t.Helper()

	root = t.TempDir()
	objects = filepath.Join(root, ".git", "annex", "objects")

	require.NoError(t, os.MkdirAll(objects, fs.ModeDir))

	return root, objects
}

// annexFile links a library name to an object, writing the object only when present is true.
func annexFile(t *testing.T, root, objects, name, key string, present bool) string {
	t.Helper()

	object := filepath.Join(objects, key)

	if present {
		require.NoError(t, os.WriteFile(object, []byte("annexed "+key), fs.ModeFile))
	}

	link := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(link), fs.ModeDir))

	rel, err := filepath.Rel(filepath.Dir(link), object)
	require.NoError(t, err)
	require.NoError(t, os.Symlink(rel, link))

	return link
}

// TestMediaFile_AnnexDestinations covers a library whose originals are symbolic links, which is what
// git-annex produces. A link is left to whoever created it, whether or not its content is present.
func TestMediaFile_AnnexDestinations(t *testing.T) {
	source := func(t *testing.T, dir string) *MediaFile {
		t.Helper()

		name := filepath.Join(dir, "incoming.jpg")
		require.NoError(t, os.WriteFile(name, []byte("payload"), fs.ModeFile))

		f, err := NewMediaFile(name)
		require.NoError(t, err)

		return f
	}

	t.Run("CopyRefusesADroppedAnnexFile", func(t *testing.T) {
		// A locked annex file whose content was dropped is a link with no target, which an existence
		// check that resolves the name cannot see.
		root, objects := annexLibrary(t)
		link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256-dropped", false)

		require.False(t, fs.Exists(link), "a dropped annex file does not resolve")
		require.True(t, fs.IsSymlink(link), "and is still a name in the library")

		for _, force := range []bool{false, true} {
			err := source(t, t.TempDir()).Copy(link, force)

			require.Error(t, err, "force=%v", force)
			assert.Contains(t, err.Error(), "symbolic link")
			assert.NoFileExists(t, filepath.Join(objects, "SHA256-dropped"), "the annex object must not be created")
		}
	})
	t.Run("CopyRefusesAPresentAnnexFile", func(t *testing.T) {
		root, objects := annexLibrary(t)
		link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256-present", true)

		for _, force := range []bool{false, true} {
			err := source(t, t.TempDir()).Copy(link, force)

			require.Error(t, err, "force=%v", force)

			b, readErr := os.ReadFile(filepath.Join(objects, "SHA256-present")) //nolint:gosec // test reads a temp file
			require.NoError(t, readErr)
			assert.Equal(t, "annexed SHA256-present", string(b), "the annex object must be left as it was")
		}
	})
	t.Run("MoveRefusesADroppedAnnexFile", func(t *testing.T) {
		root, objects := annexLibrary(t)
		link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256-dropped", false)

		for _, force := range []bool{false, true} {
			dir := t.TempDir()
			f := source(t, dir)

			err := f.Move(link, force)

			require.Error(t, err, "force=%v", force)
			assert.NoFileExists(t, filepath.Join(objects, "SHA256-dropped"), "the annex object must not be created")
			assert.FileExists(t, f.FileName(), "a refused move must keep its source")
		}
	})
	t.Run("AnnexDirectoryStillReceivesTheFile", func(t *testing.T) {
		// Only the destination name is refused. A library whose folders are links, which is the other
		// half of the supported layout, keeps working.
		root, _ := annexLibrary(t)

		real := filepath.Join(t.TempDir(), "external")
		require.NoError(t, os.MkdirAll(real, fs.ModeDir))
		require.NoError(t, os.Symlink(real, filepath.Join(root, "2031")))

		dest := filepath.Join(root, "2031", "photo.jpg")
		require.NoError(t, source(t, t.TempDir()).Copy(dest, false))

		b, err := os.ReadFile(filepath.Join(real, "photo.jpg")) //nolint:gosec // test reads a temp file
		require.NoError(t, err)
		assert.Equal(t, "payload", string(b), "the copy must land in the linked directory")
	})
}
