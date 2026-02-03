package fs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindDir(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		result := FindDir([]string{"./testdata"})
		assert.True(t, strings.HasSuffix(result, "/pkg/fs/testdata"))
	})
	t.Run("NotFound", func(t *testing.T) {
		result := FindDir([]string{"/directory", "/directory/subdirectory", "/linked"})
		assert.Equal(t, "", result)
	})
}

func TestMkdirAll(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		assert.NoError(t, MkdirAll("testdata"))
	})
}
