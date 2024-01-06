package api

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/server/header"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AuthToken returns the client authentication token from the request context,
// or an empty string if none is found.
func AuthToken(c *gin.Context) string {
	// Default is an empty string if no context or ID is set.
	if c == nil {
		return ""
	}

	// First check the X-Session-ID header for an existing ID.
	if id := clean.ID(c.GetHeader(header.SessionID)); id != "" {
		return id
	}

	// Otherwise, return the bearer token, if any.
	return BearerToken(c)
}

// BearerToken returns the client bearer token header value, or an empty string if none is found.
func BearerToken(c *gin.Context) string {
	if authType, bearerToken := Authorization(c); authType == header.BearerAuth && bearerToken != "" {
		return bearerToken
	}

	return ""
}

// Authorization returns the authentication type and token from the authorization request header,
// or an empty string if there is none.
func Authorization(c *gin.Context) (authType, authToken string) {
	if c == nil {
		return "", ""
	} else if s := c.GetHeader(header.Authorization); s == "" {
		// Ignore.
	} else if t := strings.Split(s, " "); len(t) != 2 {
		// Ignore.
	} else {
		return clean.ID(t[0]), clean.ID(t[1])
	}

	return "", ""
}

// AddRequestAuthorizationHeader adds a bearer token authorization header to a request.
func AddRequestAuthorizationHeader(r *http.Request, authToken string) {
	if authToken != "" {
		r.Header.Add(header.Authorization, fmt.Sprintf("%s %s", header.BearerAuth, authToken))
	}
}

// BasicAuth checks the basic authorization header for credentials and returns them if found.
//
// Note that OAuth 2.0 defines basic authentication differently than RFC 7617, however, this
// does not matter as long as only alphanumeric characters are used for client id and secret:
// https://www.scottbrady91.com/oauth/client-authentication#:~:text=OAuth%20Basic%20Authentication
func BasicAuth(c *gin.Context) (username, password, cacheKey string) {
	authType, authToken := Authorization(c)

	if authType != header.BasicAuth || authToken == "" {
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
