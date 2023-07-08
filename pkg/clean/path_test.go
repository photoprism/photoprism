package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	t.Run("ValidPath", func(t *testing.T) {
		assert.Equal(t, "/go/src/github.com/photoprism/photoprism", Path("/go/src/github.com/photoprism/photoprism"))
	})
	t.Run("ValidFile", func(t *testing.T) {
		assert.Equal(t, "filename.TXT", Path("filename.TXT"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "The quick brown fox.", Path("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", Path("filename.txt"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Path(""))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.Equal(t, ".", Path("."))
	})
	t.Run("DotDot", func(t *testing.T) {
		assert.Equal(t, "", Path(".."))
	})
	t.Run("DotDotDot", func(t *testing.T) {
		assert.Equal(t, "", Path("..."))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "", Path("${https://<host>:<port>/<path>}"))
	})
	t.Run("Control Character", func(t *testing.T) {
		assert.Equal(t, "filename.", Path("filename."+string(rune(127))))
	})
	t.Run("Special Chars", func(t *testing.T) {
		assert.Equal(t, "filename.", Path("filename.?**"))
	})
}

func TestUserPath(t *testing.T) {
	t.Run("ValidPath", func(t *testing.T) {
		assert.Equal(t, "go/src/github.com/photoprism/photoprism", UserPath("/go/src/github.com/photoprism/photoprism"))
	})
	t.Run("ValidFile", func(t *testing.T) {
		assert.Equal(t, "filename.TXT", UserPath("filename.TXT"))
	})
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "The quick brown fox", UserPath("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", UserPath("filename.txt"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", UserPath(""))
	})
	t.Run("Dot", func(t *testing.T) {
		assert.Equal(t, "", UserPath("."))
	})
	t.Run("DotDot", func(t *testing.T) {
		assert.Equal(t, "", UserPath(".."))
	})
	t.Run("DotDotDot", func(t *testing.T) {
		assert.Equal(t, "", UserPath("..."))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "", UserPath("${https://<host>:<port>/<path>}"))
	})
	t.Run("Unclean", func(t *testing.T) {
		assert.Equal(t, "foo/bar/baz", UserPath("/foo/bar/baz/"))
		assert.Equal(t, "dirty/path", UserPath("/dirty/path/"))
		assert.Equal(t, "dev.txt", UserPath("dev.txt"))
		assert.Equal(t, "", UserPath("../hello/foo/bar/../todo.txt"))
		assert.Equal(t, "hello/foo/bar/todo.txt", UserPath(". ./hello/foo/bar/./todo.txt"))
		assert.Equal(t, "", UserPath("./hello/foo/./bar/. ./todo.txt"))
		assert.Equal(t, "", UserPath(".."))
		assert.Equal(t, "", UserPath("."))
		assert.Equal(t, "", UserPath("/"))
		assert.Equal(t, "", UserPath(""))
	})
}
