package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachePath(t *testing.T) {
	t.Run("hash too short", func(t *testing.T) {
		result, err := CachePath("/foo/bar", "123", "baz", false)

		assert.Equal(t, "", result)
		assert.EqualError(t, err, "cache: hash '123' is too short")
	})

	t.Run("namespace empty", func(t *testing.T) {
		result, err := CachePath("/foo/bar", "123hjfju567695", "", false)

		assert.Equal(t, "", result)
		assert.EqualError(t, err, "cache: namespace for hash '123hjfju567695' is empty")
	})

	t.Run("1234567890abcdef", func(t *testing.T) {
		result, err := CachePath("/foo/bar", "1234567890abcdef", "baz", false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "/foo/bar/baz/1/2/3", result)
	})

	t.Run("create", func(t *testing.T) {
		ns := "pkg_fs_test"
		result, err := CachePath(os.TempDir(), "1234567890abcdef", ns, true)

		if err != nil {
			t.Fatal(err)
		}

		expected := filepath.Join(os.TempDir(), ns, "1", "2", "3")

		assert.Equal(t, expected, result)
		assert.DirExists(t, expected)

		_ = os.Remove(expected)
	})
}
