package api

import (
	"errors"
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
	openIdConnect := service.Oidc()

	router.GET("/auth/external", func(c *gin.Context) {
		handle := openIdConnect.AuthUrlHandler()
		handle(c.Writer, c.Request)
		//url := openIdConnect.AuthUrl()
		//log.Debugf("Step1 - Get AuthCode: %q", url)
		//c.Redirect(http.StatusFound, url)
		return
	})

	router.GET(oidc.RedirectPath, func(c *gin.Context) {
		userInfo, err := openIdConnect.CodeExchangeUserInfo(c)
		if err != nil {
			log.Errorf("%s", err)
			return
		}

		u := &entity.User{
			FullName:     userInfo.GetName(),
			UserName:     userInfo.GetPreferredUsername(),
			PrimaryEmail: userInfo.GetEmail(),
			ExternalID:   userInfo.GetSubject(),
			RoleAdmin:    true,
		}

		log.Debugf("USER:\nfn: %s\nun: %s\npe: %s\nei: %s\n", u.FullName, u.UserName, u.PrimaryEmail, u.ExternalID)
		//err = u.Validate()
		//if err != nil {
		//	CallbackError(c, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		user := entity.CreateOrUpdateExternalUser(u)
		if user == nil {
			e := errors.New("api: server error. Check backend logs")
			c.Error(e)
			CallbackError(c, e.Error(), http.StatusInternalServerError)
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

func CallbackError(c *gin.Context, err string, status int) {
	c.Abort()
	c.HTML(status, "callback.tmpl", gin.H{
		"status": "error",
		"errors": err,
	})
}
