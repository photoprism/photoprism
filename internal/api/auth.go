package api

import (
	"net/http"
	"strings"

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
		openIdConnect, _ := service.Oidc()
		if err := openIdConnect.IsAvailable(); err != nil {
			c.Error(err)
			callbackError(c, err.Error(), http.StatusInternalServerError)
			return
		}

		handle := openIdConnect.AuthUrlHandler()
		handle(c.Writer, c.Request)
		return
	})

	router.GET(oidc.RedirectPath, func(c *gin.Context) {
		openIdConnect, _ := service.Oidc()

		userInfo, err := openIdConnect.CodeExchangeUserInfo(c)
		if err != nil {
			c.Error(err)
			callbackError(c, err.Error(), http.StatusInternalServerError)
			return
		}
		var uname string
		if len(userInfo.GetPreferredUsername()) >= 4 {
			uname = userInfo.GetPreferredUsername()
		} else if len(userInfo.GetNickname()) >= 4 {
			uname = userInfo.GetNickname()
		} else if len(userInfo.GetName()) >= 4 {
			uname = strings.ReplaceAll(strings.ToLower(userInfo.GetName()), " ", "-")
		} else if len(userInfo.GetEmail()) >= 4 {
			uname = userInfo.GetEmail()
		} else {
			log.Error("auth: no username found")
		}

		u := &entity.User{
			FullName:     userInfo.GetName(),
			UserName:     uname,
			PrimaryEmail: userInfo.GetEmail(),
			ExternalID:   userInfo.GetSubject(),
		}

		log.Debugf("USER: %s %s %s %s\n", u.FullName, u.UserName, u.PrimaryEmail, u.ExternalID)

		user, e := entity.CreateOrUpdateExternalUser(u)
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
