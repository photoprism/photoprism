package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Session finds the client session for the specified
// auth token, or returns nil if not found.
func Session(authToken string) *entity.Session {
	// Skip authentication when running in public mode.
	if get.Config().Public() {
		return get.Session().Public()
	} else if !rnd.IsAuthToken(authToken) {
		return nil
	}

	// Find the session based on the hashed auth
	// token used as id, or return nil otherwise.
	if s, err := get.Session().Get(rnd.SessionID(authToken)); err != nil {
		return nil
	} else {
		return s
	}
}

// SessionResponse returns authentication response data based on the session and client config.
func SessionResponse(authToken string, sess *entity.Session, conf config.ClientConfig) gin.H {
	if authToken == "" {
		return gin.H{
			"status":     "ok",
			"id":         sess.ID,
			"expires_in": sess.ExpiresIn(),
			"provider":   sess.Provider().String(),
			"user":       sess.User(),
			"data":       sess.Data(),
			"config":     conf,
		}
	} else {
		return gin.H{
			"status":       "ok",
			"id":           sess.ID,
			"access_token": authToken,
			"token_type":   sess.AuthTokenType(),
			"expires_in":   sess.ExpiresIn(),
			"provider":     sess.Provider().String(),
			"user":         sess.User(),
			"data":         sess.Data(),
			"config":       conf,
		}
	}
}

// SessionDeleteResponse returns a confirmation response for deleted sessions.
func SessionDeleteResponse(authToken string) gin.H {
	if authToken == "" {
		return gin.H{"status": "ok"}
	} else {
		return gin.H{"status": "ok", "id": authToken, "access_token": authToken}
	}
}
