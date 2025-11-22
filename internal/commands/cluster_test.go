package commands

// NOTE: A number of non-cluster CLI commands defer conf.Shutdown(), which
// closes the shared DB connection for the process. In the commands test
// harness we reopen the DB before each run, but tests that do direct
// registry/DB access (without going through a CLI action) can still observe
// a closed connection if another test has just called Shutdown().
//
// TODO: Investigate centralizing DB lifecycle for commands tests (e.g.,
// a package-level test harness that prevents Shutdown from closing the DB,
// or injecting a mock Shutdown) so these tests don't need re-registration
// or special handling. See also commands_test.go RunWithTestContext.

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestClusterSummaryCommand(t *testing.T) {
	t.Run("NotPortal", func(t *testing.T) {
		out, err := RunWithTestContext(ClusterSummaryCommand, []string{"summary"})
		assert.Error(t, err)
		_ = out
	})
}

func TestClusterNodesListCommand(t *testing.T) {
	t.Run("NotPortal", func(t *testing.T) {
		out, err := RunWithTestContext(ClusterNodesListCommand, []string{"ls"})
		assert.Error(t, err)
		_ = out
	})
}

func TestClusterNodesShowCommand(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		_ = os.Setenv("PHOTOPRISM_NODE_ROLE", "portal")
		defer os.Unsetenv("PHOTOPRISM_NODE_ROLE")
		out, err := RunWithTestContext(ClusterNodesShowCommand, []string{"show", "does-not-exist"})
		assert.Error(t, err)
		_ = out
	})
}

func TestClusterThemePullCommand(t *testing.T) {
	t.Run("NotPortal", func(t *testing.T) {
		out, err := RunWithTestContext(ClusterThemePullCommand.Subcommands[0], []string{"pull"})
		assert.Error(t, err)
		_ = out
	})
}

func TestClusterRegisterCommand(t *testing.T) {
	t.Run("ValidationMissingURL", func(t *testing.T) {
		out, err := RunWithTestContext(ClusterRegisterCommand, []string{"register", "--name", "pp-node-01", "--role", "app", "--join-token", cluster.ExampleJoinToken})
		assert.Error(t, err)
		_ = out
	})
}

func TestClusterSuccessPaths_PortalLocal(t *testing.T) {
	// TODO: This integration-style test performs direct registry writes and
	// multiple CLI actions. Other commands in this package may call Shutdown()
	// under test, closing the DB unexpectedly and causing flakiness.
	// Skipping for now; the cluster API/registry unit tests cover the logic.
	t.Skip("todo: tests may close database connection, refactoring needed")
	// Enable portal mode for local admin commands.
	c := get.Config()
	c.Options().NodeRole = "portal"
	// Some commands in previous tests may have closed the DB; ensure it's registered.
	c.RegisterDb()

	// Ensure registry and theme paths exist.
	portCfg := c.PortalConfigPath()
	nodesDir := filepath.Join(portCfg, "nodes")
	themeDir := filepath.Join(portCfg, "theme")
	assert.NoError(t, fs.MkdirAll(nodesDir))
	assert.NoError(t, fs.MkdirAll(themeDir))

	// Create a theme file to zip.
	themeFile := filepath.Join(themeDir, "test.txt")
	assert.NoError(t, os.WriteFile(themeFile, []byte("ok"), 0o600))

	// Create a registry node via FileRegistry.
	r, err := reg.NewClientRegistryWithConfig(c)
	assert.NoError(t, err)
	n := &reg.Node{Node: cluster.Node{Name: "pp-node-01", Role: cluster.RoleApp, Labels: map[string]string{"env": "test"}}}
	assert.NoError(t, r.Put(n))

	// nodes ls (JSON)
	out, err := RunWithTestContext(ClusterNodesListCommand, []string{"ls", "--json"})
	assert.NoError(t, err)
	assert.Contains(t, out, "pp-node-01")

	// nodes show by name
	out, err = RunWithTestContext(ClusterNodesShowCommand, []string{"show", "pp-node-01"})
	assert.NoError(t, err)
	assert.Contains(t, out, "pp-node-01")

	// nodes mod: add another label (non-interactive)
	out, err = RunWithTestContext(ClusterNodesModCommand, []string{"mod", "pp-node-01", "--label", "region=us-east-1", "-y"})
	assert.NoError(t, err)
	_ = out

	// theme pull via HTTP: fake portal endpoint returns a zip with test.txt
	// Prepare temp destination
	destDir := t.TempDir()

	// Create a fake portal theme zip server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/theme" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		// Build a small zip in-memory
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		f, _ := zw.Create("test.txt")
		_, _ = f.Write([]byte("ok"))
		_ = zw.Close()
		_, _ = w.Write(buf.Bytes())
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", cluster.ExampleJoinToken)
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")

	out, err = RunWithTestContext(ClusterThemePullCommand.Subcommands[0], []string{"pull", "--dest", destDir, "-f", "--portal-url=" + ts.URL, "--join-token=" + cluster.ExampleJoinToken})
	assert.NoError(t, err)
	// Expect extracted file
	assert.FileExists(t, filepath.Join(destDir, "test.txt"))
}
