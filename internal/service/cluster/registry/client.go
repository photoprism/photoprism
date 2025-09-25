package registry

import (
	"sort"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
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
		UUID:         c.NodeUUID,
		Name:         c.ClientName,
		Role:         c.ClientRole,
		ClientID:     c.ClientUID,
		AdvertiseUrl: c.ClientURL,
		Labels:       map[string]string{},
		CreatedAt:    c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.UTC().Format(time.RFC3339),
	}
	data := c.GetData()
	if data != nil {
		if data.Labels != nil {
			n.Labels = data.Labels
		}
		n.SiteUrl = data.SiteURL
		if db := data.Database; db != nil {
			n.Database.Name = db.Name
			n.Database.User = db.User
			n.Database.Driver = db.Driver
			n.Database.RotatedAt = db.RotatedAt
		}
		n.RotatedAt = data.RotatedAt
	}
	return n
}

func (r *ClientRegistry) Put(n *Node) error {
	// Upsert client preferring NodeUUID (primary), then ClientID, then Name.
	var m *entity.Client

	// 1) Try NodeUUID first, if provided.
	if n.UUID != "" {
		var existing entity.Client
		if err := entity.UnscopedDb().Where("node_uuid = ?", n.UUID).First(&existing).Error; err == nil && existing.ClientUID != "" {
			m = &existing
		}
	}

	// 2) Fall back to ClientID if not found by UUID and ClientID is valid.
	if m == nil && rnd.IsUID(n.ClientID, entity.ClientUID) {
		if existing := entity.FindClientByUID(n.ClientID); existing != nil {
			m = existing
		}
	}

	// 3) Finally, try by Name (latest by UpdatedAt). Avoid mismatching when a UUID is provided but name belongs to another node.
	if m == nil && n.Name != "" {
		var list []entity.Client
		if err := entity.UnscopedDb().Where("client_name = ?", n.Name).Find(&list).Error; err == nil && len(list) > 0 {
			// pick latest
			latest := &list[0]
			for i := 1; i < len(list); i++ {
				if list[i].UpdatedAt.After(latest.UpdatedAt) {
					latest = &list[i]
				}
			}
			// If caller provided a UUID, do not attach to a different UUID.
			if n.UUID == "" || latest.NodeUUID == n.UUID || latest.NodeUUID == "" {
				m = latest
			}
		}
	}

	if m == nil {
		m = entity.NewClient()
	}

	// Apply fields.
	if n.Name != "" {
		m.ClientName = clean.DNSLabel(n.Name)
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
	if n.UUID != "" {
		m.NodeUUID = n.UUID
	}
	data.RotatedAt = n.RotatedAt
	if n.Database.Name != "" || n.Database.User != "" || n.Database.RotatedAt != "" {
		if data.Database == nil {
			data.Database = &entity.ClientDatabase{}
		}
		data.Database.Name = n.Database.Name
		data.Database.User = n.Database.User
		data.Database.Driver = n.Database.Driver
		data.Database.RotatedAt = n.Database.RotatedAt
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
	// (e.g., API handlers) can return the actual ClientID and timestamps.
	// Note: Do not overwrite sensitive in-memory fields like Secret.
	n.ClientID = m.ClientUID
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
			n.Database.Name = db.Name
			n.Database.User = db.User
			n.Database.RotatedAt = db.RotatedAt
		}
		n.RotatedAt = data.RotatedAt
	}
	// Set initial secret if provided on create/update.
	if n.ClientSecret != "" {
		if err := m.SetSecret(n.ClientSecret); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientRegistry) Get(id string) (*Node, error) {
	// Get by NodeUUID (UUID is primary identifier)
	if id == "" {
		return nil, ErrNotFound
	}
	var c entity.Client
	if err := entity.UnscopedDb().Where("node_uuid = ?", id).First(&c).Error; err != nil || c.ClientUID == "" {
		return nil, ErrNotFound
	}
	return toNode(&c), nil
}

func (r *ClientRegistry) FindByName(name string) (*Node, error) {
	name = clean.DNSLabel(name)
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

// FindByNodeUUID looks up a node by its NodeUUID and returns the latest record.
func (r *ClientRegistry) FindByNodeUUID(nodeUUID string) (*Node, error) {
	if nodeUUID == "" {
		return nil, ErrNotFound
	}
	var list []entity.Client
	if err := entity.UnscopedDb().Where("node_uuid = ?", nodeUUID).Find(&list).Error; err != nil {
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

// FindByClientID looks up a node by its OAuth client identifier.
func (r *ClientRegistry) FindByClientID(id string) (*Node, error) {
	if !rnd.IsUID(id, entity.ClientUID) {
		return nil, ErrNotFound
	}
	c := entity.FindClientByUID(id)
	if c == nil {
		return nil, ErrNotFound
	}
	return toNode(c), nil
}

// GetClusterNodeByUUID returns a redacted cluster.Node DTO for a given NodeUUID.
// Use NodeOptsForSession to control exposure when wiring to HTTP handlers.
func (r *ClientRegistry) GetClusterNodeByUUID(nodeUUID string, opts NodeOpts) (cluster.Node, error) {
	n, err := r.FindByNodeUUID(nodeUUID)
	if err != nil || n == nil {
		return cluster.Node{}, err
	}
	return BuildClusterNode(*n, opts), nil
}

func (r *ClientRegistry) List() ([]Node, error) {
	var list []entity.Client
	// Identify cluster nodes primarily by presence of NodeUUID.
	if err := entity.UnscopedDb().Where("node_uuid <> ''").Find(&list).Error; err != nil {
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

func (r *ClientRegistry) Delete(uuid string) error {
	if uuid == "" {
		return ErrNotFound
	}
	// Delete the latest record for this UUID (typical case: only one).
	n, err := r.FindByNodeUUID(uuid)
	if err != nil || n == nil || n.ClientID == "" {
		return ErrNotFound
	}
	c := entity.FindClientByUID(n.ClientID)
	if c == nil {
		return ErrNotFound
	}
	return c.Delete()
}

// DeleteAllByUUID removes all client rows that match the given NodeUUID.
func (r *ClientRegistry) DeleteAllByUUID(uuid string) error {
	if uuid == "" {
		return ErrNotFound
	}
	var list []entity.Client
	if err := entity.UnscopedDb().Where("node_uuid = ?", uuid).Find(&list).Error; err != nil {
		return err
	}
	if len(list) == 0 {
		return ErrNotFound
	}
	for i := range list {
		if err := list[i].Delete(); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientRegistry) RotateSecret(uuid string) (*Node, error) {
	if uuid == "" {
		return nil, ErrNotFound
	}
	n, err := r.FindByNodeUUID(uuid)
	if err != nil || n == nil || n.ClientID == "" {
		return nil, ErrNotFound
	}
	c := entity.FindClientByUID(n.ClientID)
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
	data.RotatedAt = time.Now().UTC().Format(time.RFC3339)
	c.SetData(data)
	if err := c.Save(); err != nil {
		return nil, err
	}
	n = toNode(c)
	n.ClientSecret = secret // plaintext only in-memory for response composition
	return n, nil
}
