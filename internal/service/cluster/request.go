package cluster

// RegisterRequest represents the JSON payload sent to the Portal when a node
// registers or refreshes its metadata.
//
// swagger:model RegisterRequest
type RegisterRequest struct {
	NodeName       string            `json:"nodeName"`
	NodeUUID       string            `json:"nodeUUID,omitempty"`
	NodeRole       string            `json:"nodeRole,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	AdvertiseUrl   string            `json:"advertiseUrl,omitempty"`
	SiteUrl        string            `json:"siteUrl,omitempty"`
	ClientID       string            `json:"clientId,omitempty"`
	ClientSecret   string            `json:"clientSecret,omitempty"`
	RotateDatabase bool              `json:"rotateDatabase,omitempty"`
	RotateSecret   bool              `json:"rotateSecret,omitempty"`
}
