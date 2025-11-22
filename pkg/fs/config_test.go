package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFilePath(t *testing.T) {
	pwd, _ := os.Getwd()

	t.Run("EmptyName", func(t *testing.T) {
		assert.Equal(t, "", ConfigFilePath("", "", ""))
		assert.Equal(t, "", ConfigFilePath("", "", ExtYml))
		assert.Equal(t, "", ConfigFilePath("./testdata", "", ExtYml))
	})
	t.Run("EmptyPath", func(t *testing.T) {
		assert.Equal(t, filepath.Join(pwd, "example.json"), ConfigFilePath("", "example", ExtJson))
		assert.Equal(t, filepath.Join(pwd, "example.yml"), ConfigFilePath("", "example", ExtYml))
		assert.Equal(t, filepath.Join(pwd, "example.yaml"), ConfigFilePath("", "example", ExtYaml))
	})
	t.Run("ExtNone", func(t *testing.T) {
		configPath := "testdata/config"
		envPath := filepath.Join(configPath, ".env")
		fooPath := filepath.Join(configPath, ".foo")
		fooPathLocal := fooPath + ExtLocal

		assert.Equal(t, envPath, ConfigFilePath(configPath, ".env", ExtNone))
		assert.Equal(t, fooPathLocal, ConfigFilePath(configPath, ".foo", ExtNone))
	})
	t.Run("YmlFileExists", func(t *testing.T) {
		dir := t.TempDir()
		name := "app-config"

		// Create .yml file
		ymlPath := filepath.Join(dir, name+ExtYml)
		err := os.WriteFile(ymlPath, []byte("foo: bar\n"), ModeFile)
		if err != nil {
			t.Fatalf("write %s: %v", ymlPath, err)
		}

		assert.Equal(t, ymlPath, ConfigFilePath(dir, name, ExtYml))
		assert.Equal(t, ymlPath, ConfigFilePath(dir, name, ExtYaml))
	})
	t.Run("YamlFilesMissing", func(t *testing.T) {
		dir := t.TempDir()
		name := "settings"

		// Ensure .yml does not exist; do not create it.
		ymlPath := filepath.Join(dir, name+ExtYml)
		yamlPath := filepath.Join(dir, name+ExtYaml)

		assert.Equal(t, ymlPath, ConfigFilePath(dir, name, ExtYml))
		assert.Equal(t, yamlPath, ConfigFilePath(dir, name, ExtYaml))
	})
	t.Run("BothYamlFilesExist", func(t *testing.T) {
		dir := t.TempDir()
		name := "prefs"

		// Create both files.
		ymlPath := filepath.Join(dir, name+ExtYml)
		yamlPath := filepath.Join(dir, name+ExtYaml)

		if err := os.WriteFile(ymlPath, []byte("a: 1\n"), ModeFile); err != nil {
			t.Fatalf("write %s: %v", ymlPath, err)
		}

		if err := os.WriteFile(yamlPath, []byte("a: 2\n"), ModeFile); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		assert.Equal(t, ymlPath, ConfigFilePath(dir, name, ExtYml))
		assert.Equal(t, yamlPath, ConfigFilePath(dir, name, ExtYaml))
	})
	t.Run("AlternateExtensions", func(t *testing.T) {
		tests := []struct {
			name          string
			defaultExt    string
			altExts       []string
			expectPathIdx int
		}{
			{name: "geo", defaultExt: ExtGeoJson, altExts: []string{ExtJson}},
			{name: "tml", defaultExt: ExtTml, altExts: []string{ExtToml}},
			{name: "toml", defaultExt: ExtToml, altExts: []string{ExtTml}},
			{name: "md", defaultExt: ExtMd, altExts: []string{ExtMarkdown}},
			{name: "markdown", defaultExt: ExtMarkdown, altExts: []string{ExtMd}},
			{name: "html", defaultExt: ExtHTML, altExts: []string{ExtHTM, ExtXHTML}},
			{name: "html-xhtml", defaultExt: ExtHTML, altExts: []string{ExtXHTML}, expectPathIdx: 0},
			{name: "htm", defaultExt: ExtHTM, altExts: []string{ExtHTML, ExtXHTML}},
			{name: "htm-xhtml", defaultExt: ExtHTM, altExts: []string{ExtXHTML}, expectPathIdx: 0},
			{name: "pb", defaultExt: ExtPb, altExts: []string{ExtProto}},
			{name: "proto", defaultExt: ExtProto, altExts: []string{ExtPb}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				dir := t.TempDir()
				base := "config-" + tc.name

				var paths []string
				for _, ext := range tc.altExts {
					path := filepath.Join(dir, base+ext)
					if err := os.WriteFile(path, []byte(ext+" file"), ModeFile); err != nil {
						t.Fatalf("write %s: %v", path, err)
					}
					paths = append(paths, path)
				}

				expected := paths[tc.expectPathIdx]
				got := ConfigFilePath(dir, base, tc.defaultExt)
				assert.Equal(t, expected, got)
			})
		}
	})
	t.Run("DefaultExtensionPreferred", func(t *testing.T) {
		tests := []struct {
			name       string
			defaultExt string
			altExt     string
		}{
			{name: "yaml-vs-yml", defaultExt: ExtYaml, altExt: ExtYml},
			{name: "html-vs-htm", defaultExt: ExtHTML, altExt: ExtHTM},
			{name: "proto-vs-pb", defaultExt: ExtProto, altExt: ExtPb},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				dir := t.TempDir()
				base := "prefer-" + tc.name

				defaultPath := filepath.Join(dir, base+tc.defaultExt)
				if err := os.WriteFile(defaultPath, []byte("default"), ModeFile); err != nil {
					t.Fatalf("write %s: %v", defaultPath, err)
				}

				altPath := filepath.Join(dir, base+tc.altExt)
				if err := os.WriteFile(altPath, []byte("alt"), ModeFile); err != nil {
					t.Fatalf("write %s: %v", altPath, err)
				}

				got := ConfigFilePath(dir, base, tc.defaultExt)
				assert.Equal(t, defaultPath, got)
			})
		}
	})
}
