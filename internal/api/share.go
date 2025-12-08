package api

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ShareToken creates a session using the specified share token and renders the generic sharing bootstrap page.
//
//	@Summary	creates a session using the specified share token and renders the generic sharing bootstrap page
//	@Id			ShareToken
//	@Tags		Sharing
//	@Produce	text/html
//	@Param		token	path		string	true	"Share token"
//	@Success	200		{string}	string	"Rendered HTML page"
//	@Failure	302		{string}	string	"Redirect to the base site when the token is invalid"
//	@Router		/s/{token} [get]
func ShareToken(router *gin.RouterGroup) {
	router.GET("/:token", func(c *gin.Context) {
		conf := get.Config()

		token := clean.Token(c.Param("token"))
		links := entity.FindValidLinks(token, "")

		if len(links) == 0 {
			log.Debugf("share: invalid token")
			c.Redirect(http.StatusTemporaryRedirect, conf.BaseUri(""))
			return
		}

		clientConfig := conf.ClientShare()
		clientConfig.SiteUrl += path.Join("s", token)

		uri := conf.LibraryUri("/albums")
		c.HTML(http.StatusOK, "share.gohtml", gin.H{"shared": gin.H{"token": token, "uri": uri}, "config": clientConfig})
	})
}

// ShareTokenShared creates a session with the specified share token and redirects to the shared content.
//
//	@Summary	creates a session with the specified share token and redirects to the shared content
//	@Id			ShareTokenShared
//	@Tags		Sharing
//	@Produce	text/html
//	@Param		token	path		string	true	"Share token"
//	@Param		shared	path		string	true	"Shared resource UID"
//	@Success	200		{string}	string	"Rendered HTML page"
//	@Failure	302		{string}	string	"Redirect to the base site when the token is invalid"
//	@Router		/s/{token}/{shared} [get]
func ShareTokenShared(router *gin.RouterGroup) {
	router.GET("/:token/:shared", func(c *gin.Context) {
		conf := get.Config()

		token := clean.Token(c.Param("token"))
		shared := clean.Token(c.Param("shared"))

		links := entity.FindValidLinks(token, shared)

		if len(links) < 1 {
			log.Debugf("share: invalid token or slug")
			c.Redirect(http.StatusTemporaryRedirect, conf.BaseUri(""))
			return
		}

		uid := links[0].ShareUID
		clientConfig := conf.ClientShare()
		clientConfig.SiteUrl += path.Join("s", token, uid)
		clientConfig.SitePreview = clientConfig.SiteUrl + "/preview"

		if a, err := query.AlbumByUID(uid); err == nil {
			clientConfig.SiteCaption = a.AlbumTitle

			if a.AlbumDescription != "" {
				clientConfig.SiteDescription = a.AlbumDescription
			}
		}

		uri := conf.LibraryUri(path.Join("/albums", uid, "view"))

		c.HTML(http.StatusOK, "share.gohtml", gin.H{"shared": gin.H{"token": token, "uri": uri}, "config": clientConfig})
	})
}
