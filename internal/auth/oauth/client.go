package oauth

import (
	"errors"
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

// ErrInvalidClient is returned when a client cannot use the app flow.
var ErrInvalidClient = errors.New("invalid client")

// Client represents an app client that users can connect with the authorization code flow.
type Client struct {
	ID           string
	Name         string
	Scope        string
	RedirectURIs []string
}

// FindClient returns the app client with the specified client ID, or ErrInvalidClient if the app flow is
// unavailable for it.
func FindClient(conf *config.Config, clientID string) (*Client, error) {
	if !conf.OAuthAppsEnabled() || clientID == "" {
		return nil, ErrInvalidClient
	}

	// Client ID Metadata Documents are not accepted yet.
	if strings.HasPrefix(clientID, "https://") {
		return nil, ErrInvalidClient
	}

	m := entity.FindClientByUID(clientID)

	if m == nil || m.Deleted() || m.Disabled() || !m.IsApp() {
		return nil, ErrInvalidClient
	}

	return &Client{
		ID:           m.ClientUID,
		Name:         m.Name(),
		Scope:        m.Scope(),
		RedirectURIs: m.RedirectURIs(),
	}, nil
}

// RedirectOrigin returns the origin of a redirect URI for display: "scheme://host" for web URLs,
// and the scheme for private-use URI schemes of native apps.
func RedirectOrigin(redirectURI string) string {
	u, err := url.Parse(redirectURI)

	if err != nil || u.Scheme == "" {
		return ""
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return strings.TrimSuffix(scheme.OriginURL(redirectURI), "/")
	default:
		return strings.ToLower(u.Scheme) + ":"
	}
}
