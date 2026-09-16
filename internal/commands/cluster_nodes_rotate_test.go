package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// createTestNode registers a node for the duration of a test and removes it afterwards.
func createTestNode(t *testing.T, regy *reg.ClientRegistry, name, role string) *reg.Node {
	t.Helper()

	n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: name, Role: role}}
	assert.NoError(t, regy.Create(n))
	t.Cleanup(func() { _ = regy.DeleteAllByUUID(n.UUID) })

	return n
}

func TestRotateNodeInRegistry(t *testing.T) {
	conf := get.Config()

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	t.Run("RotateSecret", func(t *testing.T) {
		n := createTestNode(t, regy, "pp-rotate-local", cluster.RoleInstance)
		before, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		resp, err := rotateNodeInRegistry(conf, "pp-rotate-local", false, true)
		assert.NoError(t, err)
		assert.Equal(t, n.UUID, resp.Node.UUID)
		assert.Equal(t, before.ClientID, resp.Node.ClientID)
		assert.True(t, resp.AlreadyRegistered)

		// The reported secret is the one now stored, and the previous one no longer matches.
		if assert.NotNil(t, resp.Secrets) {
			assert.NotEmpty(t, resp.Secrets.ClientSecret)

			client := entity.FindClientByUID(before.ClientID)

			if assert.NotNil(t, client) {
				assert.True(t, client.VerifySecret(resp.Secrets.ClientSecret))
				assert.False(t, client.VerifySecret(before.ClientSecret))
			}
		}
	})
	t.Run("RotatesTheResolvedRecord", func(t *testing.T) {
		n := createTestNode(t, regy, "pp-rotate-target", cluster.RoleInstance)
		target, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		// A second record sharing the UUID must not receive the new secret, as the rotation
		// addresses the record the name resolved to.
		other := entity.NewClient()
		other.ClientName = "pp-rotate-shadow"
		other.NodeUUID = n.UUID
		assert.NoError(t, other.Create())
		assert.NoError(t, other.SetSecret(cluster.ExampleClientSecret))
		t.Cleanup(func() { _ = other.Delete() })

		resp, err := rotateNodeInRegistry(conf, "pp-rotate-target", false, true)
		assert.NoError(t, err)
		assert.Equal(t, target.ClientID, resp.Node.ClientID)

		if assert.NotNil(t, resp.Secrets) {
			shadow := entity.FindClientByUID(other.ClientUID)

			if assert.NotNil(t, shadow) {
				assert.True(t, shadow.VerifySecret(cluster.ExampleClientSecret))
				assert.False(t, shadow.VerifySecret(resp.Secrets.ClientSecret))
			}
		}
	})
	t.Run("ReportsStoredDatabaseMetadata", func(t *testing.T) {
		n := createTestNode(t, regy, "pp-rotate-db", cluster.RoleInstance)
		n.Database = &cluster.NodeDatabase{Name: "cluster_dtest", User: "cluster_utest", Driver: "mysql", RotatedAt: "2026-09-10T00:00:00Z"}
		assert.NoError(t, regy.Put(n))

		// Rotating only the secret still reports what the record holds, as the HTTP path does.
		resp, err := rotateNodeInRegistry(conf, "pp-rotate-db", false, true)
		assert.NoError(t, err)
		assert.Equal(t, "cluster_dtest", resp.Database.Name)
		assert.Equal(t, "cluster_utest", resp.Database.User)
		assert.Equal(t, "mysql", resp.Database.Driver)
		assert.True(t, resp.AlreadyProvisioned)
		assert.Empty(t, resp.Database.Password, "a password is reported only when it is rotated")
	})
	t.Run("UnknownNode", func(t *testing.T) {
		_, err := rotateNodeInRegistry(conf, "pp-rotate-missing", false, true)

		if assert.Error(t, err) {
			var exitErr cli.ExitCoder
			assert.ErrorAs(t, err, &exitErr)
			assert.Equal(t, 3, exitErr.ExitCode())
		}
	})
	t.Run("NotANodeClient", func(t *testing.T) {
		// An ordinary OAuth client shares the name space and is not eligible.
		other := entity.NewClient()
		other.ClientName = "pp-rotate-plain"
		assert.NoError(t, other.Create())
		t.Cleanup(func() { _ = other.Delete() })

		_, err := rotateNodeInRegistry(conf, "pp-rotate-plain", false, true)

		if assert.Error(t, err) {
			var exitErr cli.ExitCoder
			assert.ErrorAs(t, err, &exitErr)
			assert.Equal(t, 3, exitErr.ExitCode())
		}
	})
	t.Run("NoRotationRequested", func(t *testing.T) {
		n := createTestNode(t, regy, "pp-rotate-none", cluster.RoleInstance)

		resp, err := rotateNodeInRegistry(conf, "pp-rotate-none", false, false)
		assert.NoError(t, err)
		assert.Nil(t, resp.Secrets)
		assert.Equal(t, n.UUID, resp.Node.UUID)

		// The operator view reports the client identifier so the table can print it.
		assert.NotEmpty(t, resp.Node.ClientID)
	})
}
