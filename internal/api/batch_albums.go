package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// BatchAlbumsDelete permanently removes multiple albums.
//
//	@Summary	permanently removes multiple albums
//	@Id			BatchAlbumsDelete
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	i18n.Response
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		albums				body		form.Selection	true	"Album Selection"
//	@Router		/api/v1/batch/albums/delete [post]
func BatchAlbumsDelete(router *gin.RouterGroup) {
	router.POST("/batch/albums/delete", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		var frm form.Selection

		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		// Get album UIDs.
		albumUIDs := frm.Albums

		if len(albumUIDs) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoAlbumsSelected)
			return
		}

		log.Infof("albums: deleting %s", clean.Log(frm.String()))

		// Fetch albums.
		albums, queryErr := query.AlbumsByUID(albumUIDs, false)

		if queryErr != nil {
			log.Errorf("albums: %s (find)", queryErr)
		}

		// Abort if no albums with a matching UID were found.
		if len(albums) == 0 {
			AbortEntityNotFound(c)
			return
		}

		deleted := 0
		conf := get.Config()

		// Flag matching albums as deleted.
		for _, a := range albums {
			if deleteErr := a.Delete(); deleteErr != nil {
				log.Errorf("albums: %s (delete)", deleteErr)
			} else {
				if conf.BackupAlbums() {
					SaveAlbumYaml(&a)
				}

				deleted++
			}
		}

		// Update client config if at least one album was successfully deleted.
		if deleted > 0 {
			UpdateClientConfig()
		}

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgAlbumsDeleted))
	})
}
