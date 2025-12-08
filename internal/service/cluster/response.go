package cluster

// NodeDatabase represents database metadata returned for a node.
// swagger:model NodeDatabase
type NodeDatabase struct {
	Name      string `json:"Name"`
	User      string `json:"User"`
	Driver    string `json:"Driver,omitempty"`
	RotatedAt string `json:"RotatedAt"`
}

// Node is the API response DTO for a cluster node.
// swagger:model Node
type Node struct {
	UUID         string            `json:"UUID"` // NodeUUID
	Name         string            `json:"Name"` // NodeName
	Role         string            `json:"Role"` // NodeRole
	ClientID     string            `json:"ClientID,omitempty"`
	AppName      string            `json:"AppName,omitempty"`
	AppVersion   string            `json:"AppVersion,omitempty"`
	Theme        string            `json:"Theme,omitempty"`
	SiteUrl      string            `json:"SiteUrl,omitempty"`
	AdvertiseUrl string            `json:"AdvertiseUrl,omitempty"`
	Labels       map[string]string `json:"Labels,omitempty"`
	CreatedAt    string            `json:"CreatedAt"`
	UpdatedAt    string            `json:"UpdatedAt"`
	Database     *NodeDatabase     `json:"Database,omitempty"`
}

// DatabaseInfo provides basic database connection metadata for summary endpoints.
// swagger:model DatabaseInfo
type DatabaseInfo struct {
	Driver string `json:"Driver"`
	Host   string `json:"Host"`
	Port   int    `json:"Port"`
}

// SummaryResponse is the response type for GET /api/v1/cluster.
// swagger:model SummaryResponse
type SummaryResponse struct {
	UUID        string       `json:"UUID"` // ClusterUUID
	ClusterCIDR string       `json:"ClusterCIDR,omitempty"`
	Nodes       int          `json:"Nodes"`
	Database    DatabaseInfo `json:"Database"`
	Theme       string       `json:"Theme,omitempty"`
	Time        string       `json:"Time"`
}

// MetricsResponse is the response type for GET /api/v1/cluster/metrics.
// swagger:model MetricsResponse
type MetricsResponse struct {
	UUID        string         `json:"UUID"`
	ClusterCIDR string         `json:"ClusterCIDR,omitempty"`
	Nodes       map[string]int `json:"Nodes"`
	Time        string         `json:"Time"`
}

// RegisterSecrets contains newly issued or rotated node secrets.
// swagger:model RegisterSecrets
type RegisterSecrets struct {
	ClientSecret string `json:"ClientSecret,omitempty"`
	RotatedAt    string `json:"RotatedAt,omitempty"`
}

// RegisterDatabase describes database credentials returned during registration/rotation.
// swagger:model RegisterDatabase
type RegisterDatabase struct {
	Driver    string `json:"Driver"`
	Host      string `json:"Host"`
	Port      int    `json:"Port"`
	Name      string `json:"Name"`
	User      string `json:"User"`
	Password  string `json:"Password,omitempty"`
	DSN       string `json:"DSN,omitempty"`
	RotatedAt string `json:"RotatedAt,omitempty"`
}

// RegisterResponse is the response body for POST /api/v1/cluster/nodes/register.
// swagger:model RegisterResponse
type RegisterResponse struct {
	UUID               string           `json:"UUID"` // ClusterUUID
	ClusterCIDR        string           `json:"ClusterCIDR,omitempty"`
	Node               Node             `json:"Node"`
	Database           RegisterDatabase `json:"Database"`
	Secrets            *RegisterSecrets `json:"Secrets,omitempty"`
	JWKSUrl            string           `json:"JWKSUrl,omitempty"`
	AlreadyRegistered  bool             `json:"AlreadyRegistered"`
	AlreadyProvisioned bool             `json:"AlreadyProvisioned"`
	Theme              string           `json:"Theme,omitempty"`
}

// StatusResponse is a generic status wrapper for simple ok responses.
// swagger:model StatusResponse
type StatusResponse struct {
	Status string `json:"Status"`
}
