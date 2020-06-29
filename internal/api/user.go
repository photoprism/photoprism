package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// PUT /api/v1/users/:uid/password
func ChangePassword(router *gin.RouterGroup) {
	router.PUT("/users/:uid/password", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePeople, acl.ActionUpdateSelf)

		if s.Invalid() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		uid := c.Param("uid")
		m := entity.FindPersonByUID(uid)

		if m == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		f := form.ChangePassword{}

		if err := c.BindJSON(&f); err != nil {
			log.Errorf("user: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrFormInvalid)
			return
		}

		if m.InvalidPassword(f.OldPassword) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrFormInvalid)
			return
		}

		if err := m.SetPassword(f.NewPassword); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "password changed"})
	})
}
