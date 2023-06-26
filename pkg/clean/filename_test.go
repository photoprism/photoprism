package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileName(t *testing.T) {
	t.Run("Path", func(t *testing.T) {
		assert.Equal(t, "", FileName("/go/src/github.com/photoprism/photoprism"))
	})
	t.Run("File", func(t *testing.T) {
		assert.Equal(t, "filename.TXT", FileName("filename.TXT"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "The quick brown fox.", FileName("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", FileName("filename.txt"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", FileName(""))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.Equal(t, "", FileName("."))
	})
	t.Run("DotDot", func(t *testing.T) {
		assert.Equal(t, "", FileName(".."))
	})
	t.Run("DotDotDot", func(t *testing.T) {
		assert.Equal(t, "", FileName("..."))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "", FileName("${https://<host>:<port>/<path>}"))
	})
	t.Run("file?name.jpg", func(t *testing.T) {
		assert.Equal(t, "filename.jpg", FileName("file?name.jpg"))
	})
	t.Run("Control Character", func(t *testing.T) {
		assert.Equal(t, "filename.", FileName("filename."+string(rune(127))))
	})
}
