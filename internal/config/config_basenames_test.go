package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestConfigStandardBasenames preserves both supported YAML suffixes across configuration files.
func TestConfigStandardBasenames(t *testing.T) {
	for _, tc := range []struct{ name, extension string }{
		{"Yml", ".yml"}, {"Yaml", ".yaml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewMinimalTestConfig(t.TempDir())
			c.options.OptionsYaml = ""
			require.NoError(t, fs.MkdirAll(c.ConfigPath()))

			for _, name := range []string{"options", "defaults", "settings", "hub"} {
				require.NoError(t, os.WriteFile(filepath.Join(c.ConfigPath(), name+tc.extension), []byte("control: true\n"), fs.ModeConfigFile))
			}

			assert.Equal(t, filepath.Join(c.ConfigPath(), "options"+tc.extension), c.OptionsYaml())
			assert.Equal(t, filepath.Join(c.ConfigPath(), "settings"+tc.extension), c.SettingsYaml())
			assert.Equal(t, filepath.Join(c.ConfigPath(), "hub"+tc.extension), c.HubConfigFile())

			ctx := CliTestContext()
			require.NoError(t, ctx.Set("config-path", c.ConfigPath()))
			require.NoError(t, ctx.Set("defaults-yaml", ""))
			c.options.DefaultsYaml = defaultsYaml(ctx)
			assert.Equal(t, filepath.Join(c.ConfigPath(), "defaults"+tc.extension), c.DefaultsYaml())
			assert.Equal(t, filepath.Join(c.ConfigPath(), "settings"+tc.extension), c.SettingsYamlDefaults(""))
		})
	}
}
