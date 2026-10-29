package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Rotating secret selects the latest row for a UUID and persists rotation timestamp and password.
func TestClientRegistry_RotateSecretByUUID_LatestRow(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-rotate-latest")

	r, _ := NewClientRegistryWithConfig(c)
	uuid := rnd.UUIDv7()

	// Two rows share the NodeUUID: a registered node and a later duplicate, created directly because a
	// second Put with the same UUID updates the first row.
	n1 := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-rot-a", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(n1))
	backdateClient(t, n1.ClientID, time.Minute)
	c2 := entity.NewClient().SetName("pp-rot-b").SetRole(cluster.RoleInstance)
	c2.NodeUUID = uuid
	require.NoError(t, c2.Create())

	// Rotate by UUID
	rotated, err := r.RotateSecret(uuid)
	assert.NoError(t, err)
	if assert.NotNil(t, rotated) {
		assert.NotEmpty(t, rotated.ClientSecret)
		assert.Equal(t, uuid, rotated.UUID)
		assert.Equal(t, c2.ClientUID, rotated.ClientID, "rotation selects the latest row")
		// Password row updated for latest ClientID
		pw := entity.FindPassword(rotated.ClientID)
		if assert.NotNil(t, pw) {
			assert.True(t, pw.Valid(rotated.ClientSecret))
		}
	}

	// Rotation timestamp persisted in client data
	got, err := r.FindByNodeUUID(uuid)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.NotEmpty(t, got.RotatedAt)
	}
}

func TestClientRegistry_RotateSecretByClientID_RevokesSessions(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-rotate-sessions")

	r, err := NewClientRegistryWithConfig(c)
	require.NoError(t, err)

	n := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-rot-sessions", Role: cluster.RoleInstance}}
	require.NoError(t, r.Put(n))

	client := entity.FindClientByUID(n.ClientID)
	require.NotNil(t, client)

	// newSession creates an access token session for the node client.
	newSession := func() string {
		sess := entity.NewSession(3600, 0)
		sess.SetClient(client)
		require.NoError(t, sess.Save())
		return sess.ID
	}

	t.Run("Omit", func(t *testing.T) {
		keep, other := newSession(), newSession()

		_, err := r.RotateSecretByClientID(n.ClientID, keep)
		require.NoError(t, err)

		_, keepErr := entity.FindSession(keep)
		assert.NoError(t, keepErr)
		_, otherErr := entity.FindSession(other)
		assert.Error(t, otherErr)
	})
	t.Run("All", func(t *testing.T) {
		ids := []string{newSession(), newSession()}

		_, err := r.RotateSecretByClientID(n.ClientID)
		require.NoError(t, err)

		for _, id := range ids {
			_, findErr := entity.FindSession(id)
			assert.Error(t, findErr)
		}
	})
}
