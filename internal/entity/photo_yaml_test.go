package entity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestPhoto_Yaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		result, err := m.Yaml()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("YAML: %s", result)
	})
}

func TestPhoto_SaveAsYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		fileName := filepath.Join(os.TempDir(), ".photoprism_test.yml")

		if err := m.SaveAsYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := m.LoadFromYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("FilenameEmpty", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		err := m.SaveAsYaml("")

		assert.Error(t, err)
	})
	t.Run("NoPhotoUID", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		fileName := filepath.Join(os.TempDir(), ".photoprism_test.yml")

		err := m.SaveAsYaml(fileName)

		assert.Error(t, err)
	})
}

func TestPhoto_YamlFileName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		op := filepath.Join(t.TempDir(), "xxx")
		fileName, relative, err := m.YamlFileName(op, "yyy")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(op, "2790/02/yyy/Photo01.yml"), fileName)
		assert.Equal(t, "2790/02/Photo01.yml", relative)
	})
}

func TestPhoto_SaveSidecarYaml(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "yaml")
	require.NoError(t, os.MkdirAll(basePath, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(basePath) })
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		require.NoError(t, fs.MkdirAll(originalsPath))
		require.NoError(t, fs.MkdirAll(sidecarPath))

		require.NoError(t, m.SaveSidecarYaml(originalsPath, sidecarPath))

		require.NoError(t, os.RemoveAll(originalsPath))
		require.NoError(t, os.RemoveAll(sidecarPath))
	})
	t.Run("PhotoNameEmpty", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		require.NoError(t, fs.MkdirAll(originalsPath))
		require.NoError(t, fs.MkdirAll(sidecarPath))

		require.Error(t, m.SaveSidecarYaml(originalsPath, sidecarPath))

		require.NoError(t, os.RemoveAll(originalsPath))
		require.NoError(t, os.RemoveAll(sidecarPath))
	})
	t.Run("PhotoUIDEmpty", func(t *testing.T) {
		m := Photo{PhotoName: "testphoto"}
		m.PreloadFiles()

		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		require.NoError(t, fs.MkdirAll(originalsPath))
		require.NoError(t, fs.MkdirAll(sidecarPath))

		require.Error(t, m.SaveSidecarYaml(originalsPath, sidecarPath))

		require.NoError(t, os.RemoveAll(originalsPath))
		require.NoError(t, os.RemoveAll(sidecarPath))
	})
}

func TestPhoto_LoadFromYaml(t *testing.T) {
	t.Run("EmptyFilename", func(t *testing.T) {
		m := Photo{}

		err := m.LoadFromYaml("")

		assert.Error(t, err)
	})
}
