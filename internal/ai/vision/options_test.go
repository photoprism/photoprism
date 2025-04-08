package vision

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestOptions(t *testing.T) {
	var configPath = fs.Abs("testdata")
	var configFile = filepath.Join(configPath, "vision.yml")

	t.Run("Save", func(t *testing.T) {
		_ = os.Remove(configFile)
		options := NewOptions()
		err := options.Save(configFile)
		assert.NoError(t, err)
		err = options.Load(configFile)
		assert.NoError(t, err)
	})
	t.Run("LoadMissingFile", func(t *testing.T) {
		options := NewOptions()
		err := options.Load(filepath.Join(configPath, "invalid.yml"))
		assert.Error(t, err)
	})
}
