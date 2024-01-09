package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/header"
)

// ClientIP returns the client IP address from the request context or a placeholder if it is unknown.
func ClientIP(c *gin.Context) (ip string) {
	return header.ClientIP(c)
}

// UserAgent returns the user agent from the request context or an empty string if it is unknown.
func UserAgent(c *gin.Context) string {
	return header.UserAgent(c)
}
