package entity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_Yaml(t *testing.T) {
	t.Run("create from fixture", func(t *testing.T) {
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
	t.Run("create from fixture", func(t *testing.T) {
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
}

func TestPhoto_YamlFileName(t *testing.T) {
	t.Run("create from fixture", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		assert.Equal(t, "xxx/2790/02/yyy/Photo01.yml", m.YamlFileName("xxx", "yyy"))

		if err := os.RemoveAll("xxx"); err != nil {
			t.Fatal(err)
		}
	})
}
