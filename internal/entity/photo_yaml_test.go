package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
		err := m.SaveAsYaml("test")

		if err != nil {
			t.Fatal(err)
		}

	})
}

func TestPhoto_LoadFromYaml(t *testing.T) {
	t.Run("create from fixture", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		err := m.LoadFromYaml("test")

		if err != nil {
			t.Fatal(err)
		}

	})
}

func TestPhoto_YamlFileName(t *testing.T) {
	t.Run("create from fixture", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		assert.Equal(t, "xxx/2790/02/yyy/Photo01.yml", m.YamlFileName("xxx", "yyy"))

	})
}
