package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileName(t *testing.T) {
	t.Run("sidecar", func(t *testing.T) {
		assert.Equal(t, ".photoprism/test.jpg", FileName("sidecar", "test.jpg"))
	})
	t.Run("import", func(t *testing.T) {
		assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/import/test.jpg", FileName("import", "test.jpg"))
	})
	t.Run("examples", func(t *testing.T) {
		assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/examples/test.jpg", FileName("examples", "test.jpg"))
	})

}
