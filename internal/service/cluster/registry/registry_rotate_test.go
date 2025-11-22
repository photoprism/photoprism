package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Rotating secret selects the latest row for a UUID and persists rotation timestamp and password.
func TestClientRegistry_RotateSecretByUUID_LatestRow(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-rotate-latest", t.TempDir())
	defer c.CloseDb()

	r, _ := NewClientRegistryWithConfig(c)
	uuid := rnd.UUIDv7()

	// Create two entries for same NodeUUID; c2 will be latest
	n1 := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-rot-a", Role: cluster.RoleApp}}
	assert.NoError(t, r.Put(n1))
	time.Sleep(1100 * time.Millisecond)
	n2 := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-rot-b", Role: cluster.RoleApp}}
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
