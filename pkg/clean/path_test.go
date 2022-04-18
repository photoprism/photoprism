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
}
