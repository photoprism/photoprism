package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// annexLibrary builds the shape a git-annex library has: every original is a relative symbolic link
// into the annex object store. Returns the library root and the object store path.
func annexLibrary(t *testing.T) (root, objects string) {
	t.Helper()

	root = t.TempDir()
	objects = filepath.Join(root, ".git", "annex", "objects")

	require.NoError(t, os.MkdirAll(objects, fs.ModeDir))

	return root, objects
}

// annexObject returns the path an object key occupies, which is nested and named for the key twice.
func annexObject(objects, key string) string {
	return filepath.Join(objects, key[:2], key[2:4], key, key)
}

// annexFile links a library name to an object. Dropping the content removes the whole key directory,
// so a dropped file is a link whose parent directory is gone too.
func annexFile(t *testing.T, root, objects, name, key string, present bool) string {
	t.Helper()

	object := annexObject(objects, key)

	if present {
		require.NoError(t, os.MkdirAll(filepath.Dir(object), fs.ModeDir))
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

	t.Run("CopyRefusesALinkWithNoTarget", func(t *testing.T) {
		// A link whose target does not resolve is what an unmounted drive or a removed object leaves
		// behind, and it is the case an existence check that resolves the name reads as free.
		for _, force := range []bool{false, true} {
			root, objects := annexLibrary(t)
			link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256E-dropped", false)

			require.False(t, fs.Exists(link), "a link with no target does not resolve")
			require.True(t, fs.IsSymlink(link), "and is still a name in the library")

			err := source(t, t.TempDir()).Copy(link, force)

			assert.Error(t, err, "force=%v", force)
			assert.Contains(t, err.Error(), "symbolic link", "force=%v", force)
			assert.NoFileExists(t, annexObject(objects, "SHA256E-dropped"), "the object must not be created")
			assert.True(t, fs.IsSymlink(link), "the link must be left in place")
		}
	})
	t.Run("CopyRefusesAPresentAnnexFile", func(t *testing.T) {
		// The corruptible case: the object is there, so a write through the link would reach it. The
		// real store keeps objects read-only, which stops that for an unprivileged process but not
		// for one running as root, as the container does by default.
		for _, force := range []bool{false, true} {
			root, objects := annexLibrary(t)
			link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256E-present", true)

			err := source(t, t.TempDir()).Copy(link, force)

			assert.Error(t, err, "force=%v", force)
			assert.Contains(t, err.Error(), "symbolic link", "force=%v", force)

			b, readErr := os.ReadFile(annexObject(objects, "SHA256E-present")) //nolint:gosec // test reads a temp file
			require.NoError(t, readErr)
			assert.Equal(t, "annexed SHA256E-present", string(b), "the object must be left as it was")
		}
	})
	t.Run("MoveRefusesALinkWithNoTarget", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			root, objects := annexLibrary(t)
			link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256E-dropped", false)

			f := source(t, t.TempDir())
			srcName := f.FileName()

			err := f.Move(link, force)

			assert.Error(t, err, "force=%v", force)
			assert.Contains(t, err.Error(), "symbolic link", "force=%v", force)
			assert.NoFileExists(t, annexObject(objects, "SHA256E-dropped"), "the object must not be created")
			assert.FileExists(t, srcName, "a refused move must keep its source")
			assert.Equal(t, srcName, f.FileName(), "and must not adopt the destination name")
		}
	})
	t.Run("MoveRefusesAPresentAnnexFile", func(t *testing.T) {
		// A move onto a live link would replace it, detaching the object from the store, so this is
		// the half a dangling-only guard would let through.
		for _, force := range []bool{false, true} {
			root, objects := annexLibrary(t)
			link := annexFile(t, root, objects, "2030/05/photo.jpg", "SHA256E-present", true)

			f := source(t, t.TempDir())
			srcName := f.FileName()

			err := f.Move(link, force)

			assert.Error(t, err, "force=%v", force)
			assert.Contains(t, err.Error(), "symbolic link", "force=%v", force)
			assert.True(t, fs.IsSymlink(link), "the link must still be a link")
			assert.FileExists(t, srcName, "a refused move must keep its source")
			assert.Equal(t, srcName, f.FileName())

			b, readErr := os.ReadFile(annexObject(objects, "SHA256E-present")) //nolint:gosec // test reads a temp file
			require.NoError(t, readErr)
			assert.Equal(t, "annexed SHA256E-present", string(b), "the object must be left as it was")
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
