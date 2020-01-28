package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/session
func CreateSession(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/session", func(c *gin.Context) {
		var f form.Login

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if !conf.CheckPassword(f.Password) {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid password"})
			return
		}

		user := gin.H{"ID": 1, "FirstName": "Admin", "LastName": "", "Role": "admin", "Email": "photoprism@localhost"}

		token := session.Create(user)

		c.Header("X-Session-Token", token)

		s := gin.H{"token": token, "user": user, "config": conf.ClientConfig()}

		c.JSON(http.StatusOK, s)
	})
}

// DELETE /api/v1/session/
func DeleteSession(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/session/:token", func(c *gin.Context) {
		token := c.Param("token")

		session.Delete(token)

		c.JSON(http.StatusOK, gin.H{"status": "ok", "token": token})
	})
}

// Returns true, if user doesn't have a valid session token
func Unauthorized(c *gin.Context, conf *config.Config) bool {
	// Always return false if site is public
	if conf.Public() {
		return false
	}

	// Get session token from HTTP header
	token := c.GetHeader("X-Session-Token")

	// Check if session token is valid
	return !session.Exists(token)
}
