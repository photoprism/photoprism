package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for YamlFilePath in yaml.go using subtests.
func TestYamlFilePath(t *testing.T) {
	t.Run("CustomPath", func(t *testing.T) {
		tmp := t.TempDir()
		rel := filepath.Join(tmp, "custom", "config.yaml")
		// Do not create the file; function should simply return Abs(customFileName).
		expected := Abs(rel)
		got := YamlFilePath("", "", rel)
		assert.Equal(t, expected, got)
	})

	t.Run("PreferYmlIfExists", func(t *testing.T) {
		dir := t.TempDir()
		name := "app-config"

		// Create .yml file
		ymlPath := filepath.Join(dir, name+ExtYml)
		err := os.WriteFile(ymlPath, []byte("foo: bar\n"), 0o644)
		if err != nil {
			t.Fatalf("write %s: %v", ymlPath, err)
		}

		got := YamlFilePath(name, dir, "")
		assert.Equal(t, ymlPath, got)
	})

	t.Run("DefaultYamlWhenYmlMissing", func(t *testing.T) {
		dir := t.TempDir()
		name := "settings"

		// Ensure .yml does not exist; do not create it.
		expected := filepath.Join(dir, name+ExtYaml)
		got := YamlFilePath(name, dir, "")
		assert.Equal(t, expected, got)
	})

	t.Run("BothExistReturnsYml", func(t *testing.T) {
		dir := t.TempDir()
		name := "prefs"

		// Create both files
		ymlPath := filepath.Join(dir, name+ExtYml)
		yamlPath := filepath.Join(dir, name+ExtYaml)

		if err := os.WriteFile(ymlPath, []byte("a: 1\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", ymlPath, err)
		}
		if err := os.WriteFile(yamlPath, []byte("a: 2\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		got := YamlFilePath(name, dir, "")
		assert.Equal(t, ymlPath, got)
	})
}
