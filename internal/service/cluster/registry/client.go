package registry

import (
	"sort"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClientRegistry implements Registry using auth_clients + passwords.
type ClientRegistry struct{ conf *config.Config }

func NewClientRegistry() *ClientRegistry { return &ClientRegistry{} }

// NewClientRegistryWithConfig returns a client-backed registry; the config is accepted for parity with file-backed init.
func NewClientRegistryWithConfig(c *config.Config) (*ClientRegistry, error) {
	return &ClientRegistry{conf: c}, nil
}

// toNode maps an auth client to the registry.Node DTO used by response builders.
func toNode(c *entity.Client) *Node {
	if c == nil {
		return nil
	}
	n := &Node{
		ID:           c.ClientUID,
		Name:         c.ClientName,
		Role:         c.ClientRole,
		CreatedAt:    c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.UTC().Format(time.RFC3339),
		AdvertiseUrl: c.ClientURL,
		Labels:       map[string]string{},
	}
	data := c.GetData()
	if data != nil {
		if data.Labels != nil {
			n.Labels = data.Labels
		}
		n.SiteUrl = data.SiteURL
		if db := data.Database; db != nil {
			n.DB.Name = db.Name
			n.DB.User = db.User
			n.DB.RotAt = db.RotatedAt
		}
		n.SecretRot = data.SecretRotatedAt
	}
	return n
}

func (r *ClientRegistry) Put(n *Node) error {
	// Upsert client by UID if provided, else by name.
	var m *entity.Client
	if rnd.IsUID(n.ID, entity.ClientUID) {
		if existing := entity.FindClientByUID(n.ID); existing != nil {
			m = existing
		}
	}
	if m == nil && n.Name != "" {
		// Try by name (latest updated wins if multiple); scan minimal for now.
		var list []entity.Client
		if err := entity.UnscopedDb().Where("client_name = ?", n.Name).Find(&list).Error; err == nil {
			var latest *entity.Client
			for i := range list {
				if latest == nil || list[i].UpdatedAt.After(latest.UpdatedAt) {
					latest = &list[i]
				}
			}
			if latest != nil {
				m = latest
			}
		}
	}
	if m == nil {
		m = entity.NewClient()
	}

	// Apply fields.
	if n.Name != "" {
		m.ClientName = clean.TypeLowerDash(n.Name)
	}
	if n.Role != "" {
		m.SetRole(n.Role)
	}
	// Ensure a default scope for node clients (instance/service) if none is set.
	// Always include "vision"; this only permits access to Vision endpoints WHEN the Portal enables them.
	if m.Scope() == "" {
		role := m.AclRole().String()
		if role == "instance" || role == "service" {
			m.SetScope("cluster vision")
		}
	}
	if n.AdvertiseUrl != "" {
		m.ClientURL = n.AdvertiseUrl
	}
	data := m.GetData()
	if data.Labels == nil {
		data.Labels = map[string]string{}
	}
	for k, v := range n.Labels {
		data.Labels[k] = v
	}
	if n.SiteUrl != "" {
		data.SiteURL = n.SiteUrl
	}
	data.SecretRotatedAt = n.SecretRot
	if n.DB.Name != "" || n.DB.User != "" || n.DB.RotAt != "" {
		if data.Database == nil {
			data.Database = &entity.ClientDatabase{}
		}
		data.Database.Name = n.DB.Name
		data.Database.User = n.DB.User
		data.Database.RotatedAt = n.DB.RotAt
	}
	m.SetData(data)

	// Persist base record.
	if m.HasUID() {
		if err := m.Save(); err != nil {
			return err
		}
	} else {
		if err := m.Create(); err != nil {
			return err
		}
	}

	// Reflect persisted values back into the provided node pointer so callers
	// (e.g., API handlers) can return the actual ID and timestamps.
	// Note: Do not overwrite sensitive in-memory fields like Secret.
	n.ID = m.ClientUID
	n.Name = m.ClientName
	n.Role = m.ClientRole
	n.AdvertiseUrl = m.ClientURL
	n.CreatedAt = m.CreatedAt.UTC().Format(time.RFC3339)
	n.UpdatedAt = m.UpdatedAt.UTC().Format(time.RFC3339)

	if data := m.GetData(); data != nil {
		// Labels and Site URL as persisted.
		if data.Labels != nil {
			n.Labels = data.Labels
		}
		n.SiteUrl = data.SiteURL
		if db := data.Database; db != nil {
			n.DB.Name = db.Name
			n.DB.User = db.User
			n.DB.RotAt = db.RotatedAt
		}
		n.SecretRot = data.SecretRotatedAt
	}
	// Set initial secret if provided on create/update.
	if n.Secret != "" {
		if err := m.SetSecret(n.Secret); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientRegistry) Get(id string) (*Node, error) {
	c := entity.FindClientByUID(id)
	if c == nil {
		return nil, ErrNotFound
	}
	return toNode(c), nil
}

func (r *ClientRegistry) FindByName(name string) (*Node, error) {
	name = clean.TypeLowerDash(name)
	if name == "" {
		return nil, ErrNotFound
	}
	var list []entity.Client
	if err := entity.UnscopedDb().Where("client_name = ?", name).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNotFound
	}
	latest := &list[0]
	for i := 1; i < len(list); i++ {
		if list[i].UpdatedAt.After(latest.UpdatedAt) {
			latest = &list[i]
		}
	}
	return toNode(latest), nil
}

func (r *ClientRegistry) List() ([]Node, error) {
	var list []entity.Client
	if err := entity.UnscopedDb().Where("client_role IN (?)", []string{"instance", "service", "portal"}).Find(&list).Error; err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].UpdatedAt.After(list[j].UpdatedAt) })
	out := make([]Node, 0, len(list))
	for i := range list {
		if n := toNode(&list[i]); n != nil {
			out = append(out, *n)
		}
	}
	return out, nil
}

func (r *ClientRegistry) Delete(id string) error {
	c := entity.FindClientByUID(id)
	if c == nil {
		return ErrNotFound
	}
	return c.Delete()
}

func (r *ClientRegistry) RotateSecret(id string) (*Node, error) {
	c := entity.FindClientByUID(id)
	if c == nil {
		return nil, ErrNotFound
	}
	// Generate and persist new secret (hashed in passwords).
	secret, err := c.NewSecret()
	if err != nil {
		return nil, err
	}
	// Update rotation timestamp in data.
	data := c.GetData()
	data.SecretRotatedAt = time.Now().UTC().Format(time.RFC3339)
	c.SetData(data)
	if err := c.Save(); err != nil {
		return nil, err
	}
	n := toNode(c)
	n.Secret = secret // plaintext only in-memory for response composition
	return n, nil
}
