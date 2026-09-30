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

// shareAliases assigns the remote destinations of one upload request.
type shareAliases struct {
	taken map[string]bool
	next  map[string]int
}

// newShareAliases returns an empty destination assignment.
func newShareAliases() *shareAliases {
	return &shareAliases{taken: make(map[string]bool), next: make(map[string]int)}
}

// Resolve returns the destination of a file below folder, adding a sequence number while the
// preferred name is already taken. It resumes at the number that name last reached, and resolves
// the related photo once, since ShareBase reads it on every call.
func (a *shareAliases) Resolve(file *entity.File, folder string) string {
	if file.Photo == nil {
		file.Photo = file.RelatedPhoto()
	}

	alias := path.Join(folder, file.ShareBase(0))
	key := strings.ToLower(alias)

	// Each name already assigned can claim one candidate, so one more than their number is free.
	for tries := 0; tries <= len(a.taken) && a.taken[strings.ToLower(alias)]; tries++ {
		a.next[key]++
		alias = path.Join(folder, file.ShareBase(a.next[key]))
	}

	return alias
}

// Keep records a destination as assigned.
func (a *shareAliases) Keep(alias string) {
	a.taken[strings.ToLower(alias)] = true
}

// UploadToService uploads files to the selected service account.
//
//	@Summary	uploads files to the selected service account
//	@Id			UploadToService
//	@Tags		Services
//	@Accept		json
//	@Produce	json
//	@Param		id						path		string	true	"service id"
//	@Success	200						{object}	entity.Files
//	@Failure	400,401,403,404,413,429	{object}	i18n.Response
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

		if webdav.SkipSyncPath(folder) || webdav.UnsafeSyncPath(folder) {
			log.Tracef("services: rejected upload folder %s", clean.Log(folder))
			AbortBadRequest(c, errors.New("destination folder is not allowed"))
			return
		}

		// Find files to share within the session's scope.
		selection := query.ShareSelection(m.ShareOriginals(), m.SyncYamlEnabled())
		files, err := query.SelectedFilesForSession(frm.Selection, selection, s)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		aliases := newShareAliases()

		selected := files[:0]

		for _, file := range files {
			if webdav.SkipSyncPath(file.FileName) {
				log.Debugf("services: skipping excluded file %s", clean.Log(file.FileName))
				continue
			}

			alias := aliases.Resolve(&file, folder)

			if webdav.SkipSyncPath(alias) || webdav.UnsafeSyncPath(alias) {
				log.Debugf("services: skipping rejected destination %s", clean.Log(alias))
				continue
			}

			aliases.Keep(alias)

			entity.FirstOrCreateFileShare(entity.NewFileShare(file.ID, m.ID, alias))

			selected = append(selected, file)
		}

		workers.RunShare(get.Config())

		c.JSON(http.StatusOK, selected)
	})
}
