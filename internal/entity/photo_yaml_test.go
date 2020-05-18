package entity

import "testing"

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
