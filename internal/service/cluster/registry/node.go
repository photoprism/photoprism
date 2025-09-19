package registry

// Node represents a registered cluster node (transport DTO inside registry package).
// It is used by both client-backed and (legacy) file-backed registries.
type Node struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Role         string            `json:"role"`
	Labels       map[string]string `json:"labels"`
	SiteUrl      string            `json:"siteUrl"`
	AdvertiseUrl string            `json:"advertiseUrl"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
	Secret       string            `json:"-"` // plaintext only when newly created/rotated in-memory
	SecretRot    string            `json:"secretRotatedAt"`
	DB           struct {
		Name  string `json:"name"`
		User  string `json:"user"`
		RotAt string `json:"rotatedAt"`
	} `json:"db"`
}
