package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestNodeClientSecretDirectoryMode checks that fallback reads leave directory permissions intact.
func TestNodeClientSecretDirectoryMode(t *testing.T) {
	for _, mode := range []os.FileMode{0o700, 0o750} {
		for _, empty := range []bool{false, true} {
			t.Setenv(FlagFileVar("NODE_CLIENT_SECRET"), "")
			conf := NewMinimalTestConfig(t.TempDir())
			file := conf.NodeClientSecretFile()
			dir := filepath.Dir(file)
			require.NoError(t, fs.MkdirAll(dir))
			conf.Options().NodeClientSecret = cluster.ExampleClientSecret
			if empty {
				require.NoError(t, os.WriteFile(file, nil, fs.ModeSecretFile))
			}
			require.NoError(t, os.Chmod(dir, mode))
			before, err := os.Stat(dir)
			require.NoError(t, err)
			require.Equal(t, cluster.ExampleClientSecret, conf.NodeClientSecret())
			after, err := os.Stat(dir)
			require.NoError(t, err)
			require.Equal(t, before.Mode(), after.Mode())
			_, err = conf.SaveNodeClientSecret(cluster.ExampleClientSecret)
			require.NoError(t, err)
			after, err = os.Stat(dir)
			require.NoError(t, err)
			require.Equal(t, before.Mode(), after.Mode())
			control := filepath.Join(dir, "secret-control")
			require.NoError(t, os.WriteFile(control, nil, fs.ModeSecretFile))
			want, err := os.Stat(control)
			require.NoError(t, err)
			got, err := os.Stat(file)
			require.NoError(t, err)
			require.Equal(t, want.Mode().Perm(), got.Mode().Perm())
			t.Logf("directory mode: %04o; secret mode: %04o", after.Mode().Perm(), got.Mode().Perm())
		}
	}
}
