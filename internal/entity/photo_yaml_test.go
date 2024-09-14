package entity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
		fileName, relative, err := m.YamlFileName("xxx", "yyy")
		assert.NoError(t, err)
		assert.Equal(t, "xxx/2790/02/yyy/Photo01.yml", fileName)
		assert.Equal(t, "2790/02/Photo01.yml", relative)

		if err := os.RemoveAll("xxx"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_SaveSidecarYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		basePath := fs.Abs("testdata/yaml")
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.SaveSidecarYaml(originalsPath, sidecarPath); err != nil {
			t.Error(err)
		}

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
	t.Run("PhotoNameEmpty", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		basePath := fs.Abs("testdata/yaml")
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveSidecarYaml(originalsPath, sidecarPath)

		assert.Error(t, err)

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
	t.Run("PhotoUIDEmpty", func(t *testing.T) {
		m := Photo{PhotoName: "testphoto"}
		m.PreloadFiles()

		basePath := fs.Abs("testdata/yaml")
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveSidecarYaml(originalsPath, sidecarPath)

		assert.Error(t, err)

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
}

func TestPhoto_LoadFromYaml(t *testing.T) {
	t.Run("EmptyFilename", func(t *testing.T) {
		m := Photo{}

		err := m.LoadFromYaml("")

		assert.Error(t, err)
	})
}
