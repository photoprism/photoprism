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

	// Create two entries for same NodeUUID; c2 will be latest
	n1 := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-rot-a", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(n1))
	time.Sleep(1100 * time.Millisecond)
	n2 := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-rot-b", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(n2))

	// Rotate by UUID
	rotated, err := r.RotateSecret(uuid)
	assert.NoError(t, err)
	if assert.NotNil(t, rotated) {
		assert.NotEmpty(t, rotated.ClientSecret)
		assert.Equal(t, uuid, rotated.UUID)
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
