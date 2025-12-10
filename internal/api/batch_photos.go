package api

import (
	"net/http"
	"path"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// BatchPhotosArchive moves multiple photos to the archive.
//
//	@Summary	moves multiple photos to the archive
//	@Id			BatchPhotosArchive
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	i18n.Response
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		photos					body		form.Selection	true	"Photo Selection"
//	@Router		/api/v1/batch/photos/archive [post]
func BatchPhotosArchive(router *gin.RouterGroup) {
	router.POST("/batch/photos/archive", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		var frm form.Selection

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if len(frm.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: archiving %s", clean.Log(frm.String()))

		if get.Config().SidecarYaml() {
			// Fetch selection from index.
			photos, err := query.SelectedPhotos(frm)

			if err != nil {
				AbortEntityNotFound(c)
				return
			}

			for _, p := range photos {
				if archiveErr := p.Archive(); archiveErr != nil {
					log.Errorf("archive: %s", archiveErr)
				} else {
					SaveSidecarYaml(p)
				}
			}
		} else if err := entity.Db().Where("photo_uid IN (?)", frm.Photos).Delete(&entity.Photo{}).Error; err != nil {
			log.Errorf("archive: failed to archive %d pictures (%s)", len(frm.Photos), err)
			AbortSaveFailed(c)
			return
		} else if err = entity.Db().Model(&entity.PhotoAlbum{}).Where("photo_uid IN (?)", frm.Photos).UpdateColumn("hidden", true).Error; err != nil {
			log.Errorf("archive: failed to flag %d pictures as hidden (%s)", len(frm.Photos), err)
		}

		// Update precalculated photo and file counts.
		entity.UpdateCountsAsync()

		// Update album, subject, and label cover thumbs.
		query.UpdateCoversAsync()

		UpdateClientConfig()

		event.EntitiesArchived("photos", frm.Photos)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionArchived))
	})
}

// BatchPhotosRestore restores multiple photos from the archive.
//
//	@Summary	restores multiple photos from the archive
//	@Id			BatchPhotosRestore
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	i18n.Response
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		photos					body		form.Selection	true	"Photo Selection"
//	@Router		/api/v1/batch/photos/restore [post]
func BatchPhotosRestore(router *gin.RouterGroup) {
	router.POST("/batch/photos/restore", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		var frm form.Selection

		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if len(frm.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: restoring %s", clean.Log(frm.String()))

		if get.Config().SidecarYaml() {
			// Fetch selection from index.
			photos, err := query.SelectedPhotos(frm)

			if err != nil {
				AbortEntityNotFound(c)
				return
			}

			for _, p := range photos {
				if err = p.Restore(); err != nil {
					log.Errorf("restore: %s", err)
				} else {
					SaveSidecarYaml(p)
				}
			}
		} else if err := entity.Db().Unscoped().Model(&entity.Photo{}).Where("photo_uid IN (?)", frm.Photos).
			UpdateColumn("deleted_at", gorm.Expr("NULL")).Error; err != nil {
			log.Errorf("restore: %s", err)
			AbortSaveFailed(c)
			return
		}

		// Update precalculated photo and file counts.
		entity.UpdateCountsAsync()

		// Update album, subject, and label cover thumbs.
		query.UpdateCoversAsync()

		UpdateClientConfig()

		event.EntitiesRestored("photos", frm.Photos)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionRestored))
	})
}

// BatchPhotosApprove approves multiple photos that are currently under review.
//
//	@Summary	approves multiple photos that are currently under review
//	@Id			BatchPhotosApprove
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	i18n.Response
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		photos				body		form.Selection	true	"Photo Selection"
//	@Router		/api/v1/batch/photos/approve [post]
func BatchPhotosApprove(router *gin.RouterGroup) {
	router.POST("/batch/photos/approve", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		var frm form.Selection

		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if len(frm.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: approving %s", clean.Log(frm.String()))

		// Fetch selection from index.
		photos, err := query.SelectedPhotos(frm)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var approved entity.Photos

		for _, p := range photos {
			if err = p.Approve(); err != nil {
				log.Errorf("approve: %s", err)
			} else {
				approved = append(approved, p)
				SaveSidecarYaml(p)
			}
		}

		UpdateClientConfig()

		event.EntitiesUpdated("photos", approved)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionApproved))
	})
}

// BatchPhotosPrivate toggles private state of multiple photos.
//
//	@Summary	toggles private state of multiple photos
//	@Id			BatchPhotosPrivate
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	i18n.Response
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		photos					body		form.Selection	true	"Photo Selection"
//	@Router		/api/v1/batch/photos/private [post]
func BatchPhotosPrivate(router *gin.RouterGroup) {
	router.POST("/batch/photos/private", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.AccessPrivate)

		if s.Abort(c) {
			return
		}

		var frm form.Selection

		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if len(frm.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: updating private flag for %s", clean.Log(frm.String()))

		if err := entity.Db().Model(entity.Photo{}).Where("photo_uid IN (?)", frm.Photos).UpdateColumn("photo_private",
			gorm.Expr("CASE WHEN photo_private > 0 THEN 0 ELSE 1 END")).Error; err != nil {
			log.Errorf("private: %s", err)
			AbortSaveFailed(c)
			return
		}

		// Update precalculated photo and file counts.
		entity.UpdateCountsAsync()

		// Fetch selection from index.
		if photos, err := query.SelectedPhotos(frm); err == nil {
			for _, p := range photos {
				SaveSidecarYaml(p)
			}

			event.EntitiesUpdated("photos", photos)
		}

		UpdateClientConfig()

		FlushCoverCache()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionProtected))
	})
}

// BatchPhotosDelete permanently removes multiple photos from the archive.
//
//	@Summary	permanently removes multiple or all photos from the archive
//	@Id			BatchPhotosDelete
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	i18n.Response
//	@Failure	400,401,403,429	{object}	i18n.Response
//	@Param		photos			body		form.Selection	true	"All or Photo Selection"
//	@Router		/api/v1/batch/photos/delete [post]
func BatchPhotosDelete(router *gin.RouterGroup) {
	router.POST("/batch/photos/delete", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Delete {
			AbortFeatureDisabled(c)
			return
		}

		var frm form.Selection

		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		deleteStart := time.Now()

		var photos entity.Photos
		var err error

		// Abort if user wants to delete all but does not have sufficient privileges.
		if frm.All && !acl.Rules.AllowAll(acl.ResourcePhotos, s.GetUserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}) {
			AbortForbidden(c)
			return
		}

		// Get selection or all archived photos if f.All is true.
		switch {
		case len(frm.Photos) == 0 && !frm.All:
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		case frm.All:
			photos, err = query.ArchivedPhotos(1000000, 0)
		default:
			photos, err = query.SelectedPhotos(frm)
		}

		// Abort if the query failed or no photos were found.
		switch {
		case err != nil:
			log.Errorf("archive: %s", err)
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		case len(photos) == 0:
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		default:
			log.Infof("archive: deleting %s", english.Plural(len(photos), "photo", "photos"))
		}

		var deleted entity.Photos

		var numFiles = 0

		// Delete photos.
		for _, p := range photos {
			// Report file deletion.
			event.AuditWarn([]string{ClientIP(c), s.UserName, "delete", path.Join(p.PhotoPath, p.PhotoName+"*")})

			// Remove all related files from storage.
			n, deleteErr := photoprism.DeletePhoto(p, true, true)

			numFiles += n

			if deleteErr != nil {
				log.Errorf("delete: %s", deleteErr)
			} else {
				deleted = append(deleted, p)
			}
		}

		if numFiles > 0 || len(deleted) > 0 {
			log.Infof("archive: deleted %s and %s [%s]", english.Plural(numFiles, "file", "files"), english.Plural(len(deleted), "photo", "photos"), time.Since(deleteStart))
		}

		// Any photos deleted?
		if len(deleted) > 0 {
			config.FlushUsageCache()

			// Update precalculated photo and file counts.
			entity.UpdateCountsAsync()

			UpdateClientConfig()

			event.EntitiesDeleted("photos", deleted.UIDs())
		}

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPermanentlyDeleted))
	})
}
