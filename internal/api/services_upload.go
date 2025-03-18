package api

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UploadToService uploads files to the selected service account.
//
//	@Summary	uploads files to the selected service account
//	@Id			UploadToService
//	@Tags		Services
//	@Accept		json
//	@Produce	json
//	@Param		id				path		string	true	"service id"
//	@Success	200				{object}	entity.Files
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/services/{id}/upload [post]
func UploadToService(router *gin.RouterGroup) {
	router.POST("/services/:id/upload", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionUpload)

		if s.Abort(c) {
			return
		}

		id := clean.IdUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		var frm form.SyncUpload

		// Assign and validate request form values.
		if err = c.BindJSON(&frm); err != nil {
			AbortBadRequest(c)
			return
		}

		folder := frm.Folder

		// Find files to share.
		selection := query.ShareSelection(m.ShareOriginals())
		files, err := query.SelectedFiles(frm.Selection, selection)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var aliases = make(map[string]int)

		for _, file := range files {
			alias := path.Join(folder, file.ShareBase(0))
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.ShareBase(seq)
			}

			aliases[key] += 1

			entity.FirstOrCreateFileShare(entity.NewFileShare(file.ID, m.ID, alias))
		}

		workers.RunShare(get.Config())

		c.JSON(http.StatusOK, files)
	})
}
