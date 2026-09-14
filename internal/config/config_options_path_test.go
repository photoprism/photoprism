package config

import (
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// unwritableOptionsConfig returns a config whose options file cannot be written, together
// with the path it names. A directory stands in for the file, so the write fails for any
// process owner rather than only for one without permission.
func unwritableOptionsConfig(t *testing.T) (*Config, string) {
	t.Helper()

	dataPath := t.TempDir()
	c := NewMinimalTestConfig(dataPath)

	optionsPath := filepath.Join(dataPath, "options.yml")
	require.NoError(t, os.MkdirAll(optionsPath, fs.ModeDir))
	c.Options().OptionsYaml = optionsPath

	return c, optionsPath
}

func TestSaveOptionsPatchErrorNamesPathStructurally(t *testing.T) {
	t.Run("WriteFailure", func(t *testing.T) {
		c, optionsPath := unwritableOptionsConfig(t)

		_, err := c.SaveOptionsPatch(Values{"NodeName": "test-node"})
		require.Error(t, err)

		var pathErr *iofs.PathError
		require.ErrorAs(t, err, &pathErr)
		assert.Equal(t, optionsPath, pathErr.Path)
		assert.NotContains(t, clean.Error(err), optionsPath)
	})
	t.Run("ParseFailure", func(t *testing.T) {
		dataPath := t.TempDir()
		c := NewMinimalTestConfig(dataPath)

		optionsPath := filepath.Join(dataPath, "options.yml")
		require.NoError(t, os.WriteFile(optionsPath, []byte("\tnot: [valid"), fs.ModeConfigFile))
		c.Options().OptionsYaml = optionsPath

		_, err := c.SaveOptionsPatch(Values{"NodeName": "test-node"})
		require.Error(t, err)

		var pathErr *iofs.PathError
		require.ErrorAs(t, err, &pathErr)
		assert.Equal(t, optionsPath, pathErr.Path)
		assert.NotContains(t, clean.Error(err), optionsPath)
	})
}

func TestSetFaceModelSaveFailure(t *testing.T) {
	c, optionsPath := unwritableOptionsConfig(t)

	err := c.SetFaceModel(face.ModelSFace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), optionsPath)
	assert.NotContains(t, clean.Error(err), optionsPath)
	assert.NotContains(t, clean.Error(err), filepath.Dir(optionsPath))
}

func TestNodeUUIDSaveFailure(t *testing.T) {
	c, optionsPath := unwritableOptionsConfig(t)
	c.Options().NodeUUID = ""

	hook := captureLog(t)

	uuid := c.NodeUUID()
	assert.True(t, rnd.IsUUID(uuid))

	var warned bool

	for _, entry := range hook.AllEntries() {
		if entry.Level != logrus.WarnLevel {
			continue
		}
		warned = true
		assert.NotContains(t, entry.Message, optionsPath)
		assert.NotContains(t, entry.Message, filepath.Dir(optionsPath))
	}

	assert.True(t, warned, "expected a warning for the failed node UUID write")
}
