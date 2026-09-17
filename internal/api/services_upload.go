package api

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/webdav"
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
		LimitRequestBodyBytes(c, MaxSelectionRequestBytes)

		if err = c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		folder := frm.Folder

		if webdav.SkipSyncPath(folder) {
			log.Tracef("services: excluded upload folder %s", clean.Log(folder))
			AbortBadRequest(c, errors.New("destination folder is not allowed"))
			return
		}

		// Find files to share within the session's scope.
		selection := query.ShareSelection(m.ShareOriginals())
		files, err := query.SelectedFilesForSession(frm.Selection, selection, s)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var aliases = make(map[string]int)

		selected := files[:0]

		for _, file := range files {
			if webdav.SkipSyncPath(file.FileName) {
				log.Debugf("services: skipping excluded file %s", clean.Log(file.FileName))
				continue
			}

			alias := path.Join(folder, file.ShareBase(0))

			if webdav.SkipSyncPath(alias) {
				log.Debugf("services: skipping excluded destination %s", clean.Log(alias))
				continue
			}

			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.ShareBase(seq)
			}

			aliases[key]++

			entity.FirstOrCreateFileShare(entity.NewFileShare(file.ID, m.ID, alias))

			selected = append(selected, file)
		}

		workers.RunShare(get.Config())

		c.JSON(http.StatusOK, selected)
	})
}
