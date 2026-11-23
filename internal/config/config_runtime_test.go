package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig_Runtime(t *testing.T) {
	files := runtimeFiles
	t.Cleanup(func() { runtimeFiles = files })

	dir := t.TempDir()
	dockerFile, podmanFile := filepath.Join(dir, ".dockerenv"), filepath.Join(dir, ".containerenv")
	require.NoError(t, os.WriteFile(dockerFile, nil, fs.ModeFile))
	require.NoError(t, os.WriteFile(podmanFile, nil, fs.ModeFile))
	folder := filepath.Join(dir, "folder")
	require.NoError(t, os.Mkdir(folder, fs.ModeDir))
	missing := filepath.Join(dir, "missing")
	c := NewMinimalTestConfig(t.TempDir())

	for _, tc := range []struct {
		name       string
		kubernetes string
		docker     string
		podman     string
		container  string
		invocation string
		expected   string
	}{
		{"Host", "", missing, missing, "", "", "host"},
		{"Systemd", "", missing, missing, "", "5f7c3a9e0b2d4c6e8a1b3d5f7c9e0a2b", "systemd"},
		{"Container", "", missing, missing, "true", "", "container"},
		{"ContainerDisabled", "", missing, missing, "false", "", "host"},
		{"Podman", "", missing, podmanFile, "true", "", "podman"},
		{"Docker", "", dockerFile, podmanFile, "true", "", "docker"},
		{"Kubernetes", "10.96.0.1", dockerFile, podmanFile, "true", "", "kubernetes"},
		{"DockerBeforeSystemd", "", dockerFile, missing, "", "5f7c3a9e0b2d4c6e8a1b3d5f7c9e0a2b", "docker"},
		{"PodmanBeforeSystemd", "", missing, podmanFile, "", "5f7c3a9e0b2d4c6e8a1b3d5f7c9e0a2b", "podman"},
		{"ContainerBeforeSystemd", "", missing, missing, "true", "5f7c3a9e0b2d4c6e8a1b3d5f7c9e0a2b", "container"},
		{"FolderIsNoMarker", "", folder, folder, "", "", "host"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("KUBERNETES_SERVICE_HOST", tc.kubernetes)
			t.Setenv("PHOTOPRISM_CONTAINER", tc.container)
			t.Setenv("INVOCATION_ID", tc.invocation)
			runtimeFiles.Docker, runtimeFiles.Podman = tc.docker, tc.podman
			assert.Equal(t, tc.expected, c.Runtime())
		})
	}
}
