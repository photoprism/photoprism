package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFileName(t *testing.T) {
	conf := config.TestConfig()
	t.Run("sidecar", func(t *testing.T) {
		assert.Equal(t, ".photoprism/test.jpg", FileName("sidecar", "test.jpg"))
	})
	t.Run("import", func(t *testing.T) {
		assert.Equal(t, conf.ImportPath()+"/test.jpg", FileName("import", "test.jpg"))
	})
	t.Run("examples", func(t *testing.T) {
		assert.Equal(t, conf.ExamplesPath()+"/test.jpg", FileName("examples", "test.jpg"))
	})

}
