package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClientRegistry_PutFindListRotate(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-client")
	defer c.CloseDb()
	if err := c.Init(); err != nil {
		t.Fatalf("init config: %v", err)
	}

	r, err := NewClientRegistryWithConfig(c)
	assert.NoError(t, err)

	// Create new node
	n := &Node{
		Name:         "pp-node-a",
		Role:         "instance",
		Labels:       map[string]string{"env": "test"},
		AdvertiseUrl: "http://pp-node-a:2342",
		SiteUrl:      "https://photos.example.com",
	}
	n.DB.Name = "pp_db"
	n.DB.User = "pp_user"
	n.DB.RotAt = time.Now().UTC().Format(time.RFC3339)
	n.SecretRot = time.Now().UTC().Format(time.RFC3339)
	n.Secret = rnd.ClientSecret()

	assert.NoError(t, r.Put(n))

	// Find by name
	got, err := r.FindByName("pp-node-a")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.NotEmpty(t, got.ID)
		assert.Equal(t, "pp-node-a", got.Name)
		assert.Equal(t, "instance", got.Role)
		assert.Equal(t, "http://pp-node-a:2342", got.AdvertiseUrl)
		assert.Equal(t, "https://photos.example.com", got.SiteUrl)
		assert.Equal(t, "pp_db", got.DB.Name)
		assert.Equal(t, "pp_user", got.DB.User)
		assert.NotEmpty(t, got.CreatedAt)
		assert.NotEmpty(t, got.UpdatedAt)
		// Secret is not persisted in plaintext
		assert.Equal(t, "", got.Secret)
		assert.NotEmpty(t, got.SecretRot)
		// Password row exists and validates the initial secret
		pw := entity.FindPassword(got.ID)
		if assert.NotNil(t, pw) {
			assert.True(t, pw.Valid(n.Secret))
		}
	}

	// List contains our node
	list, err := r.List()
	assert.NoError(t, err)
	found := false
	for _, it := range list {
		if it.Name == "pp-node-a" {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Rotate secret
	rotated, err := r.RotateSecret(got.ID)
	assert.NoError(t, err)
	if assert.NotNil(t, rotated) {
		assert.NotEmpty(t, rotated.Secret)
		// Validate new secret
		pw := entity.FindPassword(got.ID)
		if assert.NotNil(t, pw) {
			assert.True(t, pw.Valid(rotated.Secret))
		}
	}

	// Update labels and site URL via Put (upsert by id)
	upd := &Node{ID: got.ID, Name: got.Name, Labels: map[string]string{"env": "prod"}, SiteUrl: "https://photos.example.org"}
	assert.NoError(t, r.Put(upd))
	got2, err := r.FindByName("pp-node-a")
	assert.NoError(t, err)
	if assert.NotNil(t, got2) {
		assert.Equal(t, "prod", got2.Labels["env"])
		assert.Equal(t, "https://photos.example.org", got2.SiteUrl)
	}
}
