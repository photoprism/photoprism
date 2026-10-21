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
	Metadata     bool
}

// ClientLookup controls how FindClient uses client metadata documents.
type ClientLookup struct {
	AllowFetch func() bool // AllowFetch, if set, is asked before a document is fetched.
	AllowStale bool        // AllowStale permits an expired document while it cannot be fetched.
}

// FindClient returns the app client with the specified client ID, or ErrInvalidClient if the app flow is
// unavailable for it. Registered clients are read through the client cache, and client metadata documents
// are looked up as specified by the optional ClientLookup.
func FindClient(conf *config.Config, clientID string, lookup ...ClientLookup) (*Client, error) {
	if !conf.OAuthAppsEnabled() || clientID == "" {
		return nil, ErrInvalidClient
	}

	// A client ID that is an HTTPS URL names a client metadata document, if the configuration allows it.
	if strings.HasPrefix(clientID, "https://") {
		if !conf.OAuthAppsMetadata(clientID) {
			return nil, ErrInvalidClient
		}

		var opt ClientLookup

		if len(lookup) > 0 {
			opt = lookup[0]
		}

		return FindMetadataClient(clientID, opt)
	}

	m, err := entity.CachedClientByUID(clientID)

	if err != nil || m.Deleted() || m.Disabled() || !m.IsApp() {
		return nil, ErrInvalidClient
	}

	return &Client{
		ID:           m.ClientUID,
		Name:         m.Name(),
		Scope:        m.Scope(),
		RedirectURIs: m.RedirectURIs(),
	}, nil
}

// ClientOrigin returns the origin shown for a client: the host of a client metadata document URL, or the
// origin of the redirect URI of a registered client.
func ClientOrigin(clientID, redirectURI string) string {
	if strings.HasPrefix(clientID, "https://") {
		return strings.TrimSuffix(scheme.OriginURL(clientID), "/")
	}

	return RedirectOrigin(redirectURI)
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
