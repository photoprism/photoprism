package vision

import (
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

// Service represents a remote computer vision service configuration.
type Service struct {
	Uri            string    `yaml:"Uri,omitempty" json:"uri"`
	Method         string    `yaml:"Method,omitempty" json:"method"`
	Key            string    `yaml:"Key,omitempty" json:"-"`
	FileScheme     string    `yaml:"FileScheme,omitempty" json:"fileScheme,omitempty"`
	RequestFormat  ApiFormat `yaml:"RequestFormat,omitempty" json:"requestFormat,omitempty"`
	ResponseFormat ApiFormat `yaml:"ResponseFormat,omitempty" json:"responseFormat,omitempty"`
	Disabled       bool      `yaml:"Disabled,omitempty" json:"disabled,omitempty"`
}

// Endpoint returns the remote service request method and endpoint URL, if any.
func (m *Service) Endpoint() (uri, method string) {
	if m.Disabled || m.Uri == "" {
		return "", ""
	}

	if m.Method != "" {
		method = m.Method
	} else {
		method = ServiceMethod
	}

	return m.Uri, method
}

// EndpointKey returns the access token belonging to the remote service endpoint, if any.
func (m *Service) EndpointKey() string {
	if m.Disabled {
		return ""
	}

	return m.Key
}

// EndpointFileScheme returns the endpoint API file scheme type.
func (m *Service) EndpointFileScheme() scheme.Type {
	if m.Disabled {
		return ""
	} else if m.FileScheme == "" {
		return ServiceFileScheme
	}

	return m.FileScheme
}

// EndpointRequestFormat returns the endpoint API request format.
func (m *Service) EndpointRequestFormat() ApiFormat {
	if m.Disabled {
		return ""
	} else if m.RequestFormat == "" {
		return ApiFormatVision
	}

	return m.RequestFormat
}

// EndpointResponseFormat returns the endpoint API response format.
func (m *Service) EndpointResponseFormat() ApiFormat {
	if m.Disabled {
		return ""
	} else if m.ResponseFormat == "" {
		return ApiFormatVision
	}

	return m.ResponseFormat
}
