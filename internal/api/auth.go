package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/oidc"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/session"
)

// GET /api/v1/auth/
func AuthEndpoints(router *gin.RouterGroup) {
	conf := service.Config()
	if conf.OidcIssuerUrl().String() == "" || conf.OidcClientId() == "" || conf.OidcClientSecret() == "" {
		log.Debugf("no oidc provider configured. skip mounting endpoints")
		return
	}
	_, err := service.Oidc()
	if err != nil {
		log.Error(err)
	}

	router.GET("/auth/external", func(c *gin.Context) {
		openIdConnect, err := service.Oidc()
		if err != nil {
			c.Error(err)
			callbackError(c, err.Error(), http.StatusInternalServerError)
			return
		}
		openIdConnect.AuthCodeUrlHandler(c)
	})

	router.GET(oidc.RedirectPath, func(c *gin.Context) {
		openIdConnect, _ := service.Oidc()

		userInfo, claimErr := openIdConnect.CodeExchangeUserInfo(c)
		if claimErr != nil {
			c.Error(claimErr)
			callbackError(c, claimErr.Error(), http.StatusInternalServerError)
			return
		}

		u := &entity.User{
			FullName:     userInfo.GetName(),
			UserName:     oidc.UsernameFromUserInfo(userInfo),
			PrimaryEmail: userInfo.GetEmail(),
			ExternalID:   userInfo.GetSubject(),
		}

		isAdmin, claimErr := oidc.HasRoleAdmin(userInfo)
		if claimErr == nil {
			u.RoleAdmin = isAdmin
			log.Debug("photoprism_admin: ", isAdmin)
		} else {
			log.Debug(claimErr)
		}

		log.Debugf("USER: %s %s %s %s\n", u.FullName, u.UserName, u.PrimaryEmail, u.ExternalID)

		user, e := entity.CreateOrUpdateExternalUser(u, claimErr == nil)
		if e != nil {
			c.Error(e)
			callbackError(c, e.Error(), http.StatusInternalServerError)
			return
		}
		log.Infof("user '%s' logged in", user.UserName)
		var data = session.Data{
			User: *user,
		}
		id := service.Session().Create(data)
		c.HTML(http.StatusOK, "callback.tmpl", gin.H{
			"status": "ok",
			"id":     id,
			"data":   data,
			"config": conf.UserConfig(),
		})

	})
}

func callbackError(c *gin.Context, err string, status int) {
	c.Abort()
	c.HTML(status, "callback.tmpl", gin.H{
		"status": "error",
		"errors": err,
	})
}
