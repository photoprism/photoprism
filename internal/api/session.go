package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/service"
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

		data := session.Data{}

		if f.HasToken() {
			links := entity.FindLinks(f.Token, "")

			if len(links) == 0 {
				c.AbortWithStatusJSON(400, gin.H{"error": "Invalid link"})
			}

			data.Tokens = []string{f.Token}

			for _, link := range links {
				data.Shared = append(data.Shared, link.ShareUID)
			}

			data.User = entity.Guest
		} else if f.HasCredentials() {
			user := entity.FindPersonByUserName(f.UserName)

			if user == nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "Invalid user name or password"})
				return
			}

			if user.InvalidPassword(f.Password) {
				c.AbortWithStatusJSON(400, gin.H{"error": "Invalid user name or password"})
				return
			}

			data.User = *user
		} else {
			c.AbortWithStatusJSON(400, gin.H{"error": "Password required, please try again"})
			return
		}

		token := service.Session().Create(data)

		c.Header("X-Session-Token", token)

		if data.User.Anonymous() {
			c.JSON(http.StatusOK, gin.H{"token": token, "data": data, "config": conf.GuestConfig()})
		} else {
			c.JSON(http.StatusOK, gin.H{"token": token, "data": data, "config": conf.UserConfig()})
		}
	})
}

// DELETE /api/v1/session/
func DeleteSession(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/session/:token", func(c *gin.Context) {
		token := c.Param("token")

		service.Session().Delete(token)

		c.JSON(http.StatusOK, gin.H{"status": "ok", "token": token})
	})
}

// Returns true, if user doesn't have a valid session token
func Unauthorized(c *gin.Context, conf *config.Config) bool {
	// Always return false if site is public.
	if conf.Public() {
		return false
	}

	// Get session token from HTTP header.
	token := c.GetHeader("X-Session-Token")

	// Check if session token is valid.
	return !service.Session().Exists(token)
}

// Gets session token from HTTP header.
func SessionToken(c *gin.Context) string {
	return c.GetHeader("X-Session-Token")
}

// Session returns the current session data.
func Session(token string, conf *config.Config) (data *session.Data) {
	if token == "" {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			data = nil
			log.Errorf("session: %s [panic]", err)
		}
	}()

	// Always return false if site is public.
	if conf.Public() {
		admin := entity.FindPersonByUserName("admin")

		if admin == nil {
			log.Error("session: admin user not found - bug?")
			return nil
		}

		return &session.Data{User: *admin}
	}

	// Check if session token is valid.
	return service.Session().Get(token)
}

// InvalidToken returns true if the token is invalid.
func InvalidToken(c *gin.Context, conf *config.Config) bool {
	token := c.Param("token")

	if token == "" {
		token = c.Query("t")
	}

	return conf.InvalidToken(token)
}

// InvalidDownloadToken returns true if the token is invalid.
func InvalidDownloadToken(c *gin.Context, conf *config.Config) bool {
	return conf.InvalidDownloadToken(c.Query("t"))
}
