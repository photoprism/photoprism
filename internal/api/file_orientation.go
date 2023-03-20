package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ChangeFileOrientation changes the orientation of a file.
// PUT /api/v1/photos/:uid/files/:file_uid/orientation
//
// Parameters:
//
//	uid: string Photo UID as returned by the API
//	file_uid: string File UID as returned by the API
func ChangeFileOrientation(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/files/:file_uid/orientation", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		// Abort in read-only mode or if editing is disabled.
		if conf.ReadOnly() || !conf.Settings().Features.Edit {
			Abort(c, http.StatusForbidden, i18n.ErrReadOnly)
			return
		} else if conf.DisableExifTool() {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "exiftool is disabled")
			return
		}

		fileUid := clean.UID(c.Param("file_uid"))

		m, err := query.FileByUID(fileUid)

		// Abort if the file was not found.
		if err != nil {
			log.Errorf("files: %s (change orientation)", err)
			AbortEntityNotFound(c)
			return
		}

		// Init form with model values
		f, err := form.NewFile(m)

		if err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		// Update form with values from request
		if err = c.BindJSON(&f); err != nil {
			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		// Update orientation if it was changed.
		if m.Orientation() != f.Orientation() {
			fileName := photoprism.FileName(m.FileRoot, m.FileName)
			mf, err := photoprism.NewMediaFile(fileName)

			// Check if file exists.
			if err != nil {
				Abort(c, http.StatusInternalServerError, i18n.ErrFileNotFound)
				return
			}

			// Update file header.
			if err = mf.ChangeOrientation(f.Orientation()); err != nil {
				log.Debugf("file: %s in %s (change orientation)", err, clean.Log(mf.BaseName()))
				Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
				return
			}

			// Update index.
			ind := get.Index()
			if res := ind.FileName(mf.FileName(), photoprism.IndexOptionsSingle()); res.Failed() {
				log.Errorf("file: %s in %s (change orientation)", res.Err, clean.Log(mf.BaseName()))
				AbortSaveFailed(c)
				return
			}
		}

		// Return updated photo.
		p, err := query.PhotoPreloadByUID(m.PhotoUID)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		PublishPhotoEvent(EntityUpdated, m.PhotoUID, c)

		c.JSON(http.StatusOK, p)
	})
}
