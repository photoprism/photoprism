package api

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UploadToService uploads files to the selected account.
//
// GET /api/v1/services/:id/upload
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

		var f form.SyncUpload

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		folder := f.Folder

		// Find files to share.
		selection := query.ShareSelection(m.ShareOriginals())
		files, err := query.SelectedFiles(f.Selection, selection)

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
