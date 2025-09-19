package cluster

// NodeDatabase represents database metadata returned for a node.
// swagger:model NodeDatabase
type NodeDatabase struct {
	Name      string `json:"name"`
	User      string `json:"user"`
	RotatedAt string `json:"rotatedAt"`
}

// Node is the API response DTO for a cluster node.
// swagger:model Node
type Node struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Role         string            `json:"role"`
	SiteUrl      string            `json:"siteUrl,omitempty"`
	AdvertiseUrl string            `json:"advertiseUrl,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
	Database     *NodeDatabase     `json:"database,omitempty"`
}

// DatabaseInfo provides basic database connection metadata for summary endpoints.
// swagger:model DatabaseInfo
type DatabaseInfo struct {
	Driver string `json:"driver"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}

// SummaryResponse is the response type for GET /api/v1/cluster.
// swagger:model SummaryResponse
type SummaryResponse struct {
	UUID     string       `json:"UUID"`
	Nodes    int          `json:"nodes"`
	Database DatabaseInfo `json:"database"`
	Time     string       `json:"time"`
}

// RegisterSecrets contains newly issued or rotated node secrets.
// swagger:model RegisterSecrets
type RegisterSecrets struct {
	NodeSecret      string `json:"nodeSecret,omitempty"`
	SecretRotatedAt string `json:"secretRotatedAt,omitempty"`
}

// RegisterDatabase describes database credentials returned during registration/rotation.
// swagger:model RegisterDatabase
type RegisterDatabase struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Name      string `json:"name"`
	User      string `json:"user"`
	Password  string `json:"password,omitempty"`
	DSN       string `json:"dsn,omitempty"`
	RotatedAt string `json:"rotatedAt,omitempty"`
}

// RegisterResponse is the response body for POST /api/v1/cluster/nodes/register.
// swagger:model RegisterResponse
type RegisterResponse struct {
	Node               Node             `json:"node"`
	Database           RegisterDatabase `json:"database"`
	Secrets            *RegisterSecrets `json:"secrets,omitempty"`
	AlreadyRegistered  bool             `json:"alreadyRegistered"`
	AlreadyProvisioned bool             `json:"alreadyProvisioned"`
}

// StatusResponse is a generic status wrapper for simple ok responses.
// swagger:model StatusResponse
type StatusResponse struct {
	Status string `json:"status"`
}
