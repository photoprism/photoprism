package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirs(t *testing.T) {
	t.Run("recursive", func(t *testing.T) {
		result, err := Dirs("testdata", true)

		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"/directory", "/directory/subdirectory", "/linked"}

		assert.Equal(t, expected, result)
	})

	t.Run("non-recursive", func(t *testing.T) {
		result, err := Dirs("testdata", false)

		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"/directory", "/linked"}

		assert.Equal(t, expected, result)
	})
}
