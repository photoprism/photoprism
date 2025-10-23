package theme

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDetectVersion(t *testing.T) {
	t.Run("VersionFilePreferred", func(t *testing.T) {
		dir := t.TempDir()
		assert.NoError(t, os.WriteFile(filepath.Join(dir, fs.VersionTxtFile), []byte(" 1.2.3 \n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, fs.AppJsFile), []byte("// app"), fs.ModeFile))

		got, err := DetectVersion(dir)
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", got)
	})
	t.Run("FallsBackToAppJS", func(t *testing.T) {
		dir := t.TempDir()
		appPath := filepath.Join(dir, fs.AppJsFile)
		assert.NoError(t, os.WriteFile(appPath, []byte("// app"), fs.ModeFile))

		want := time.Now().UTC().Truncate(time.Second)
		assert.NoError(t, os.Chtimes(appPath, want, want))

		got, err := DetectVersion(dir)
		assert.NoError(t, err)
		assert.Equal(t, want.Format(time.RFC3339), got)
	})
	t.Run("MissingAppJS", func(t *testing.T) {
		dir := t.TempDir()
		_, err := DetectVersion(dir)
		assert.Error(t, err)
	})
}
