package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
)

type CreateSessionParams struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

// POST /api/v1/session
func CreateSession(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/session", func(c *gin.Context) {
		var params CreateSessionParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if params.Password != conf.AdminPassword() {
			c.AbortWithStatusJSON(400, gin.H{"error": "Invalid password"})
			return
		}

		token, _ := util.RandomToken(16)

		c.Header("X-Session-Token", token)

		gc := conf.Cache()

		gc.Set(token, 1, cache.DefaultExpiration);

		s := gin.H{"token": token, "user": gin.H{"ID": 1, "FirstName": "Admin", "LastName": "", "Role": "admin", "Email": "photoprism@localhost"}}

		c.JSON(http.StatusOK, s)
	})
}

// DELETE /api/v1/session/
func DeleteSession(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/session/:token", func(c *gin.Context) {
		token := c.Param("token")

		gc := conf.Cache()

		gc.Delete(token)

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
	log.Debugf("X-Session-Token: %s", token)

	// Check if session token is valid
	gc := conf.Cache()
	_, found := gc.Get(token)

	return !found
}
