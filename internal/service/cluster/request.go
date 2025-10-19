package cluster

// RegisterRequest represents the JSON payload sent to the Portal when a node
// registers or refreshes its metadata.
//
// swagger:model RegisterRequest
type RegisterRequest struct {
	NodeName       string            `json:"NodeName"`
	NodeUUID       string            `json:"NodeUUID,omitempty"`
	NodeRole       string            `json:"NodeRole,omitempty"`
	Labels         map[string]string `json:"Labels,omitempty"`
	AdvertiseUrl   string            `json:"AdvertiseUrl,omitempty"`
	SiteUrl        string            `json:"SiteUrl,omitempty"`
	ClientID       string            `json:"ClientID,omitempty"`
	ClientSecret   string            `json:"ClientSecret,omitempty"`
	RotateDatabase bool              `json:"RotateDatabase,omitempty"`
	RotateSecret   bool              `json:"RotateSecret,omitempty"`
}
