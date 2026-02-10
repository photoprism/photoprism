package dirs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirs(t *testing.T) {
	t.Run("Recursive", func(t *testing.T) {
		result, counts, err := Dirs("../fs/testdata", true, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, result, 10)
		assert.Contains(t, result, "/directory")
		assert.Contains(t, result, "/directory/subdirectory")
		assert.Contains(t, result, "/directory/subdirectory/animals")
		assert.Contains(t, result, "/originals")
		assert.NotContains(t, result, "/originals/storage")
		assert.Contains(t, result, "/linked")
		assert.Equal(t, 0, counts["/directory/subdirectory/animals"])
		assert.Equal(t, 2, counts["/directory/subdirectory"])
		assert.Equal(t, 3, counts["/directory"])
	})
	t.Run("RecursiveNoSymlinks", func(t *testing.T) {
		result, _, err := Dirs("../fs/testdata", true, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, result, "/directory")
		assert.Contains(t, result, "/directory/subdirectory")
		assert.Contains(t, result, "/directory/subdirectory/animals")
		assert.Contains(t, result, "/linked")
	})
	t.Run("NonRecursive", func(t *testing.T) {
		result, counts, err := Dirs("../fs/testdata", false, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, result, "/directory")
		assert.Contains(t, result, "/linked")
		assert.Equal(t, 4, counts["/"])
		assert.Equal(t, 0, counts["/config"])
	})
	t.Run("NonRecursiveNoSymlinks", func(t *testing.T) {
		result, counts, err := Dirs("../fs/testdata/directory/subdirectory", false, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, result, 2)
		assert.Contains(t, result, "/animals")
		assert.Equal(t, 2, counts["/"])
		assert.Equal(t, 0, counts["/animals"])
	})
	t.Run("NonRecursiveSymlinks", func(t *testing.T) {
		result, _, err := Dirs("../fs/testdata/linked", false, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, result, "/photoprism")
		assert.Contains(t, result, "/self")
	})
	t.Run("NoRecursionNoChildren", func(t *testing.T) {
		result, counts, err := Dirs("../fs/testdata/linked", false, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, result, 1)
		assert.Equal(t, 0, counts["/"]) // There are only Sidecar files in folder
	})
}
