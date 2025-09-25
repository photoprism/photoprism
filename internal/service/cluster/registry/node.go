package registry

// Node represents a registered cluster node (transport DTO inside registry package).
// It is used by both client-backed and (legacy) file-backed registries.
type Node struct {
	UUID         string            `json:"uuid"` // primary identifier (UUID v7)
	Name         string            `json:"name"`
	Role         string            `json:"role"`
	ClientID     string            `json:"clientId,omitempty"` // OAuth client identifier (legacy)
	ClientSecret string            `json:"-"`                  // plaintext only when newly created/rotated in-memory
	SiteUrl      string            `json:"siteUrl,omitempty"`
	AdvertiseUrl string            `json:"advertiseUrl,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	RotatedAt    string            `json:"rotatedAt,omitempty"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
	Database     struct {
		Name      string `json:"name"`
		User      string `json:"user"`
		Driver    string `json:"driver,omitempty"`
		RotatedAt string `json:"rotatedAt,omitempty"`
	} `json:"database,omitempty"`
}
