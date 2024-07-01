package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
)

// CreateSessionResponse returns the authentication response data for POST requests
// based on the session and configuration.
func CreateSessionResponse(authToken string, sess *entity.Session, conf config.ClientConfig) gin.H {
	return GetSessionResponse(authToken, sess, conf)
}

// CreateSessionError returns an authentication error response.
func CreateSessionError(code int, err error) gin.H {
	return gin.H{
		"status": StatusFailed,
		"code":   code,
		"error":  err.Error(),
		"config": get.Config().ClientPublic(),
	}
}

// GetSessionResponse returns the authentication response data for GET requests
// based on the session and configuration.
func GetSessionResponse(authToken string, sess *entity.Session, conf config.ClientConfig) gin.H {
	if authToken == "" {
		return gin.H{
			"status":     StatusSuccess,
			"session_id": sess.ID,
			"expires_in": sess.ExpiresIn(),
			"provider":   sess.Provider().String(),
			"scope":      sess.Scope(),
			"user":       sess.User(),
			"data":       sess.Data(),
			"config":     conf,
		}
	} else {
		return gin.H{
			"status": StatusSuccess,
			// TODO: "id" field is deprecated! Clients should now use "access_token" instead.
			// see https://github.com/photoprism/photoprism/commit/0d2f8be522dbf0a051ae6ef78abfc9efded0082d
			"id":           authToken,
			"session_id":   sess.ID,
			"access_token": authToken,
			"token_type":   sess.AuthTokenType(),
			"expires_in":   sess.ExpiresIn(),
			"provider":     sess.Provider().String(),
			"scope":        sess.Scope(),
			"user":         sess.User(),
			"data":         sess.Data(),
			"config":       conf,
		}
	}
}

// DeleteSessionResponse returns a confirmation response for DELETE requests.
func DeleteSessionResponse(id string) gin.H {
	if id == "" {
		return gin.H{"status": StatusDeleted}
	} else {
		return gin.H{"status": StatusDeleted, "session_id": id}
	}
}
