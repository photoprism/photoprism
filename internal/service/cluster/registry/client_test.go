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

func TestClientRegistry_PutFindListRotate(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-client", t.TempDir())
	defer c.CloseDb()

	r, err := NewClientRegistryWithConfig(c)
	assert.NoError(t, err)

	// Create new node
	n := &Node{
		Node: cluster.Node{
			UUID:         rnd.UUIDv7(),
			Name:         "pp-node-a",
			Role:         "instance",
			AppName:      "PhotoPrism",
			AppVersion:   "1.0.0",
			Theme:        "theme-v1",
			SiteUrl:      "https://photos.example.com",
			AdvertiseUrl: "http://pp-node-a:2342",
			Labels:       map[string]string{"env": "test"},
		},
	}
	db := n.ensureDatabase()
	db.Name = "pp_db"
	db.User = "pp_user"
	db.RotatedAt = time.Now().UTC().Format(time.RFC3339)
	n.RotatedAt = time.Now().UTC().Format(time.RFC3339)
	n.ClientSecret = rnd.ClientSecret()

	assert.NoError(t, r.Put(n))

	// Find by name
	got, err := r.FindByName("pp-node-a")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.NotEmpty(t, got.ClientID)
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
		assert.True(t, rnd.IsUUID(got.UUID))
		assert.Equal(t, "pp-node-a", got.Name)
		assert.Equal(t, "instance", got.Role)
		assert.Equal(t, "PhotoPrism", got.AppName)
		assert.Equal(t, "1.0.0", got.AppVersion)
		assert.Equal(t, "theme-v1", got.Theme)
		assert.Equal(t, "http://pp-node-a:2342", got.AdvertiseUrl)
		assert.Equal(t, "https://photos.example.com", got.SiteUrl)
		if assert.NotNil(t, got.Database) {
			assert.Equal(t, "pp_db", got.Database.Name)
			assert.Equal(t, "pp_user", got.Database.User)
		}
		assert.NotEmpty(t, got.CreatedAt)
		assert.NotEmpty(t, got.UpdatedAt)
		// Secret is not persisted in plaintext
		assert.Equal(t, "", got.ClientSecret)
		assert.NotEmpty(t, got.RotatedAt)
		// Password row exists and validates the initial secret
		pw := entity.FindPassword(got.ClientID)
		if assert.NotNil(t, pw) {
			assert.True(t, pw.Valid(n.ClientSecret))
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
	rotated, err := r.RotateSecret(got.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, rotated) {
		assert.NotEmpty(t, rotated.ClientSecret)
		// Validate new secret
		pw := entity.FindPassword(got.ClientID)
		if assert.NotNil(t, pw) {
			assert.True(t, pw.Valid(rotated.ClientSecret))
		}
	}

	// Update labels and site URL via Put (upsert by id)
	upd := &Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, Labels: map[string]string{"env": "prod"}, SiteUrl: "https://photos.example.org", AppVersion: "1.1.0"}}
	assert.NoError(t, r.Put(upd))
	got2, err := r.FindByName("pp-node-a")
	assert.NoError(t, err)
	if assert.NotNil(t, got2) {
		assert.Equal(t, "prod", got2.Labels["env"])
		assert.Equal(t, "https://photos.example.org", got2.SiteUrl)
		assert.Equal(t, "PhotoPrism", got2.AppName)
		assert.Equal(t, "1.1.0", got2.AppVersion)
		assert.Equal(t, "theme-v1", got2.Theme)
	}
}
