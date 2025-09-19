package entity

import (
	"encoding/json"
)

// ClientDatabase captures DB metadata provisioned for a node.
type ClientDatabase struct {
	Name      string `json:"name,omitempty"`
	User      string `json:"user,omitempty"`
	RotatedAt string `json:"rotatedAt,omitempty"`
}

// ClientData represents instance/service-specific metadata for cluster clients.
type ClientData struct {
	Labels          map[string]string `json:"labels,omitempty"`
	Database        *ClientDatabase   `json:"database,omitempty"`
	SecretRotatedAt string            `json:"secretRotatedAt,omitempty"`
	SiteURL         string            `json:"siteUrl,omitempty"`
	ClusterUUID     string            `json:"clusterUUID,omitempty"`
	ServiceKind     string            `json:"serviceKind,omitempty"`
	ServiceFeatures []string          `json:"serviceFeatures,omitempty"`
}

// NewClientData creates a new client data struct and returns a pointer to it.
func NewClientData() *ClientData {
	return &ClientData{}
}

// GetData returns the data that belong to this session.
func (m *Client) GetData() (data *ClientData) {
	if m.data != nil {
		data = m.data
	}

	data = NewClientData()

	if len(m.DataJSON) == 0 {
		return data
	} else if err := json.Unmarshal(m.DataJSON, data); err != nil {
		log.Errorf("auth: failed to read client data (%s)", err)
	} else {
		m.data = data
	}

	return data
}

// SetData updates the data that belong to this session.
func (m *Client) SetData(data *ClientData) *Client {
	if data == nil {
		log.Debugf("auth: nil cannot be set as client data (%s)", m.ClientUID)
		return m
	}

	if j, err := json.Marshal(data); err != nil {
		log.Debugf("auth: failed to set client data (%s)", err)
	} else {
		m.DataJSON = j
	}

	m.data = data

	return m
}
