package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/session"
)

// POST /api/v1/session
func CreateSession(router *gin.RouterGroup) {
	router.POST("/session", func(c *gin.Context) {
		var f form.Login

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		var data session.Data

		id := SessionID(c)

		if s := Session(id); s.Valid() {
			data = s
		} else {
			data = session.Data{}
			id = ""
		}

		conf := service.Config()

		if f.HasToken() {
			links := entity.FindValidLinks(f.Token, "")

			if len(links) == 0 {
				c.AbortWithStatusJSON(400, gin.H{"error": i18n.Msg(i18n.ErrInvalidLink)})
			}

			data.Tokens = []string{f.Token}

			for _, link := range links {
				data.Shares = append(data.Shares, link.ShareUID)
				link.Redeem()
			}

			// Upgrade from anonymous to guest. Don't downgrade.
			if data.User.Anonymous() {
				data.User = entity.Guest
			}
		} else if f.HasCredentials() {
			user := entity.FindUserByName(f.UserName)

			if user == nil {
				c.AbortWithStatusJSON(400, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
				return
			}

			if user.InvalidPassword(f.Password) {
				c.AbortWithStatusJSON(400, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
				return
			}

			data.User = *user
		} else {
			c.AbortWithStatusJSON(400, gin.H{"error": i18n.Msg(i18n.ErrInvalidPassword)})
			return
		}

		if err := service.Session().Update(id, data); err != nil {
			id = service.Session().Create(data)
		}

		AddSessionHeader(c, id)

		if data.User.Anonymous() {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id, "data": data, "config": conf.GuestConfig()})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id, "data": data, "config": conf.UserConfig()})
		}
	})
}

// DELETE /api/v1/session/:id
func DeleteSession(router *gin.RouterGroup) {
	router.DELETE("/session/:id", func(c *gin.Context) {
		id := clean.Token(c.Param("id"))

		service.Session().Delete(id)

		c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id})
	})
}

// Gets session id from HTTP header.
func SessionID(c *gin.Context) string {
	return c.GetHeader("X-Session-ID")
}

// Session returns the current session data.
func Session(id string) session.Data {
	// Return fake admin session if site is public.
	if service.Config().Public() {
		return session.Data{User: entity.Admin}
	}

	// Check if session id is valid.
	return service.Session().Get(id)
}

// Auth returns the session if user is authorized for the current action.
func Auth(id string, resource acl.Resource, action acl.Action) session.Data {
	sess := Session(id)

	if acl.Permissions.Deny(resource, sess.User.Role(), action) {
		return session.Data{}
	}

	return sess
}

// InvalidPreviewToken returns true if the token is invalid.
func InvalidPreviewToken(c *gin.Context) bool {
	token := clean.Token(c.Param("token"))

	if token == "" {
		token = clean.Token(c.Query("t"))
	}

	return service.Config().InvalidPreviewToken(token)
}

// InvalidDownloadToken returns true if the token is invalid.
func InvalidDownloadToken(c *gin.Context) bool {
	return service.Config().InvalidDownloadToken(clean.Token(c.Query("t")))
}
