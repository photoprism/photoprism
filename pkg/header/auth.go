package header

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/pkg/clean"
)

const (
	Auth       = "Authorization" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
	AuthBasic  = "Basic"
	AuthBearer = "Bearer"
	XAuthToken = "X-Auth-Token"
	XSessionID = "X-Session-ID"
)

// AuthToken returns the client authentication token from the request context,
// or an empty string if none is found.
func AuthToken(c *gin.Context) string {
	// Default is an empty string if no context or ID is set.
	if c == nil {
		return ""
	}

	// First check the "X-Auth-Token" and "X-Session-ID" headers for an auth token.
	if token := c.GetHeader(XAuthToken); token != "" {
		return clean.ID(token)
	} else if id := c.GetHeader(XSessionID); id != "" {
		return clean.ID(id)
	}

	// Otherwise, the bearer token from the authorization request header is returned.
	return BearerToken(c)
}

// BearerToken returns the client bearer token header value, or an empty string if none is found.
func BearerToken(c *gin.Context) string {
	if authType, bearerToken := Authorization(c); authType == AuthBearer && bearerToken != "" {
		return bearerToken
	}

	return ""
}

// Authorization returns the authentication type and token from the authorization request header,
// or an empty string if there is none.
func Authorization(c *gin.Context) (authType, authToken string) {
	if c == nil {
		return "", ""
	} else if s := c.GetHeader(Auth); s == "" {
		// Ignore.
	} else if t := strings.Split(s, " "); len(t) != 2 {
		// Ignore.
	} else {
		return clean.ID(t[0]), clean.ID(t[1])
	}

	return "", ""
}

// SetAuthorization adds a bearer token authorization header to a request.
func SetAuthorization(r *http.Request, authToken string) {
	if authToken != "" {
		r.Header.Add(Auth, fmt.Sprintf("%s %s", AuthBearer, authToken))
	}
}

// BasicAuth checks the basic authorization header for credentials and returns them if found.
//
// Note that OAuth 2.0 defines basic authentication differently than RFC 7617, however, this
// does not matter as long as only alphanumeric characters are used for client id and secret:
// https://www.scottbrady91.com/oauth/client-authentication#:~:text=OAuth%20Basic%20Authentication
func BasicAuth(c *gin.Context) (username, password, cacheKey string) {
	authType, authToken := Authorization(c)

	if authType != AuthBasic || authToken == "" {
		return "", "", ""
	}

	auth, err := base64.StdEncoding.DecodeString(authToken)

	if err != nil {
		return "", "", ""
	}

	credentials := strings.SplitN(string(auth), ":", 2)

	if len(credentials) != 2 {
		return "", "", ""
	}

	cacheKey = fmt.Sprintf("%x", sha1.Sum([]byte(authToken)))

	return credentials[0], credentials[1], cacheKey
}
