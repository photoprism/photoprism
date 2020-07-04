package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/service"
)

// PUT /api/v1/users/:uid/password
func ChangePassword(router *gin.RouterGroup) {
	router.PUT("/users/:uid/password", func(c *gin.Context) {
		conf := service.Config()

		if conf.Public() {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrPublic)
			return
		}

		s := Auth(SessionID(c), acl.ResourcePeople, acl.ActionUpdateSelf)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		uid := c.Param("uid")
		m := entity.FindPersonByUID(uid)

		if m == nil {
			log.Errorf("change password: user not found")
			c.AbortWithStatusJSON(http.StatusNotFound, ErrInvalidPassword)
			return
		}

		f := form.ChangePassword{}

		if err := c.BindJSON(&f); err != nil {
			log.Errorf("change password: %s", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrInvalidPassword)
			return
		}

		if m.InvalidPassword(f.OldPassword) {
			log.Errorf("change password: invalid password")
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrInvalidPassword)
			return
		}

		if err := m.SetPassword(f.NewPassword); err != nil {
			log.Errorf("change password: %s", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "password changed"})
	})
}
