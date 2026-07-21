package duf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMounts(t *testing.T) {
	// findMounts stats and resolves the target path, so the mountpoints must be real
	// directories; build them under a temp root and evaluate symlinks up front so they
	// match the path returned by filepath.EvalSymlinks inside findMounts.
	base, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	mediaDir := filepath.Join(base, "media")
	mediaDataDir := filepath.Join(base, "media-data")
	nestedDir := filepath.Join(mediaDir, "photos")
	require.NoError(t, os.MkdirAll(nestedDir, 0o750))
	require.NoError(t, os.MkdirAll(mediaDataDir, 0o750))

	mounts := []Mount{
		{Device: "/dev/root", Mountpoint: base},
		{Device: "/dev/media", Mountpoint: mediaDir},
	}

	t.Run("Root", func(t *testing.T) {
		// A path on the root mount resolves to it.
		result, err := findMounts(mounts, base)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, base, result[0].Mountpoint)
	})
	t.Run("NestedLongestPrefix", func(t *testing.T) {
		// A nested path resolves to the closest (longest) matching mountpoint.
		result, err := findMounts(mounts, nestedDir)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, mediaDir, result[0].Mountpoint)
	})
	t.Run("ExactDevice", func(t *testing.T) {
		// A path equal to a mount's device returns that mount directly.
		devMounts := []Mount{{Device: mediaDir, Mountpoint: base}}
		result, err := findMounts(devMounts, mediaDir)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, mediaDir, result[0].Device)
	})
	t.Run("NotFound", func(t *testing.T) {
		// A path that does not exist cannot be resolved.
		result, err := findMounts(mounts, filepath.Join(base, "does-not-exist"))
		require.Error(t, err)
		assert.Nil(t, result)
	})
	t.Run("PrefixBoundary", func(t *testing.T) {
		// Regression: /media-data is a sibling of the /media mount, not a child, so it
		// must resolve to the root mount rather than matching /media on a bare prefix.
		result, err := findMounts(mounts, mediaDataDir)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, base, result[0].Mountpoint, "sibling path must not match the /media mount")
	})
}

func TestMountContains(t *testing.T) {
	sep := string(filepath.Separator)
	root := sep
	media := filepath.Join(sep, "media")
	t.Run("SamePath", func(t *testing.T) {
		assert.True(t, mountContains(media, media))
	})
	t.Run("Child", func(t *testing.T) {
		assert.True(t, mountContains(media, filepath.Join(media, "photos")))
	})
	t.Run("RootContainsEverything", func(t *testing.T) {
		assert.True(t, mountContains(root, filepath.Join(sep, "media-data")))
	})
	t.Run("SiblingNotContained", func(t *testing.T) {
		assert.False(t, mountContains(media, filepath.Join(sep, "media-data")))
	})
	t.Run("Unrelated", func(t *testing.T) {
		assert.False(t, mountContains(media, filepath.Join(sep, "var", "lib")))
	})
}
