package vision

import (
	"net/url"
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

// Service represents a remote computer vision service configuration.
type Service struct {
	Uri            string    `yaml:"Uri,omitempty" json:"uri"`
	Method         string    `yaml:"Method,omitempty" json:"method"`
	Model          string    `yaml:"Model,omitempty" json:"model,omitempty"` // Optional endpoint-specific model override.
	Username       string    `yaml:"Username,omitempty" json:"-"`            // Optional basic auth user injected into Endpoint URLs.
	Password       string    `yaml:"Password,omitempty" json:"-"`
	Key            string    `yaml:"Key,omitempty" json:"-"`
	Org            string    `yaml:"Org,omitempty" json:"org,omitempty"`         // Optional organization header (e.g. OpenAI).
	Project        string    `yaml:"Project,omitempty" json:"project,omitempty"` // Optional project header (e.g. OpenAI).
	FileScheme     string    `yaml:"FileScheme,omitempty" json:"fileScheme,omitempty"`
	RequestFormat  ApiFormat `yaml:"RequestFormat,omitempty" json:"requestFormat,omitempty"`
	ResponseFormat ApiFormat `yaml:"ResponseFormat,omitempty" json:"responseFormat,omitempty"`
	Disabled       bool      `yaml:"Disabled,omitempty" json:"disabled,omitempty"`
}

// Endpoint returns the remote service request method and endpoint URL, if any.
func (m *Service) Endpoint() (uri, method string) {
	if m.Disabled || strings.TrimSpace(m.Uri) == "" {
		return "", ""
	}

	if m.Method != "" {
		method = m.Method
	} else {
		method = ServiceMethod
	}

	uri = strings.TrimSpace(m.Uri)

	if username, password := m.BasicAuth(); username != "" || password != "" {
		if parsed, err := url.Parse(uri); err == nil {
			if parsed.User == nil {
				switch {
				case username != "" && password != "":
					parsed.User = url.UserPassword(username, password)
				case username != "":
					parsed.User = url.User(username)
				}

				if parsed.User != nil {
					uri = parsed.String()
				}
			}
		}
	}

	return uri, method
}

// GetModel returns the model identifier override for the endpoint, if any.
func (m *Service) GetModel() string {
	if m.Disabled {
		return ""
	}

	ensureEnv()

	return clean.TypeLower(os.ExpandEnv(m.Model))
}

// EndpointKey returns the access token belonging to the remote service endpoint, if any.
func (m *Service) EndpointKey() string {
	if m.Disabled {
		return ""
	}

	ensureEnv()

	return strings.TrimSpace(os.ExpandEnv(m.Key))
}

// EndpointOrg returns the organization identifier for the endpoint, if any.
func (m *Service) EndpointOrg() string {
	if m.Disabled {
		return ""
	}

	ensureEnv()

	return strings.TrimSpace(os.ExpandEnv(m.Org))
}

// EndpointProject returns the project identifier for the endpoint, if any.
func (m *Service) EndpointProject() string {
	if m.Disabled {
		return ""
	}

	ensureEnv()

	return strings.TrimSpace(os.ExpandEnv(m.Project))
}

// BasicAuth returns the username and password for basic authentication.
func (m *Service) BasicAuth() (username, password string) {
	ensureEnv()
	username = strings.TrimSpace(os.ExpandEnv(m.Username))
	password = strings.TrimSpace(os.ExpandEnv(m.Password))
	return username, password
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
