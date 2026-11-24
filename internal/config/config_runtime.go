package config

import (
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Runtime environments returned by Config.Runtime.
const (
	RuntimeKubernetes = "kubernetes"
	RuntimeDocker     = "docker"
	RuntimePodman     = "podman"
	RuntimeContainer  = "container"
	RuntimeSystemd    = "systemd"
	RuntimeHost       = "host"
)

// runtimeFiles are the marker files that Docker and Podman create in a container.
var runtimeFiles = struct{ Docker, Podman string }{Docker: "/.dockerenv", Podman: "/run/.containerenv"}

// Runtime returns the environment the process runs in: Kubernetes, which sets KUBERNETES_SERVICE_HOST in every
// pod, a container created by Docker or Podman, another container started from our images, a systemd service,
// or otherwise the host.
func (c *Config) Runtime() string {
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "":
		return RuntimeKubernetes
	case fs.FileExists(runtimeFiles.Docker):
		return RuntimeDocker
	case fs.FileExists(runtimeFiles.Podman):
		return RuntimePodman
	case txt.Bool(os.Getenv(EnvVar("CONTAINER"))):
		return RuntimeContainer
	case os.Getenv("INVOCATION_ID") != "":
		return RuntimeSystemd
	default:
		return RuntimeHost
	}
}
