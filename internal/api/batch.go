package api

import (
	"net/http"
	"path"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
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

		var f form.Selection

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if len(f.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: archiving %s", clean.Log(f.String()))

		if get.Config().SidecarYaml() {
			// Fetch selection from index.
			photos, err := query.SelectedPhotos(f)

			if err != nil {
				AbortEntityNotFound(c)
				return
			}

			for _, p := range photos {
				if archiveErr := p.Archive(); archiveErr != nil {
					log.Errorf("archive: %s", archiveErr)
				} else {
					SaveSidecarYaml(&p)
				}
			}
		} else if err := entity.Db().Where("photo_uid IN (?)", f.Photos).Delete(&entity.Photo{}).Error; err != nil {
			log.Errorf("archive: failed to archive %d pictures (%s)", len(f.Photos), err)
			AbortSaveFailed(c)
			return
		} else if err = entity.Db().Model(&entity.PhotoAlbum{}).Where("photo_uid IN (?)", f.Photos).UpdateColumn("hidden", true).Error; err != nil {
			log.Errorf("archive: failed to flag %d pictures as hidden (%s)", len(f.Photos), err)
		}

		// Update precalculated photo and file counts.
		logWarn("index", entity.UpdateCounts())

		// Update album, subject, and label cover thumbs.
		logWarn("index", query.UpdateCovers())

		UpdateClientConfig()

		event.EntitiesArchived("photos", f.Photos)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionArchived))
	})
}

// BatchPhotosRestore restores multiple photos from the archive.
//
//	@Summary	restores multiple photos from the archive
//	@Id			BatchPhotosRestore
//	@Tags		Photos
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

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if len(f.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: restoring %s", clean.Log(f.String()))

		if get.Config().SidecarYaml() {
			// Fetch selection from index.
			photos, err := query.SelectedPhotos(f)

			if err != nil {
				AbortEntityNotFound(c)
				return
			}

			for _, p := range photos {
				if err = p.Restore(); err != nil {
					log.Errorf("restore: %s", err)
				} else {
					SaveSidecarYaml(&p)
				}
			}
		} else if err := entity.Db().Unscoped().Model(&entity.Photo{}).Where("photo_uid IN (?)", f.Photos).
			UpdateColumn("deleted_at", gorm.Expr("NULL")).Error; err != nil {
			log.Errorf("restore: %s", err)
			AbortSaveFailed(c)
			return
		}

		// Update precalculated photo and file counts.
		logWarn("index", entity.UpdateCounts())

		// Update album, subject, and label cover thumbs.
		logWarn("index", query.UpdateCovers())

		UpdateClientConfig()

		event.EntitiesRestored("photos", f.Photos)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionRestored))
	})
}

// BatchPhotosApprove approves multiple photos that are currently under review.
//
//	@Summary	approves multiple photos that are currently under review
//	@Id			BatchPhotosApprove
//	@Tags		Photos
//	@Produce	json
//	@Success	200					{object}	i18n.Response
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		photos				body		form.Selection	true	"Photo Selection"
//	@Router		/api/v1/batch/photos/approve [post]
func BatchPhotosApprove(router *gin.RouterGroup) {
	router.POST("batch/photos/approve", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if len(f.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: approving %s", clean.Log(f.String()))

		// Fetch selection from index.
		photos, err := query.SelectedPhotos(f)

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
				SaveSidecarYaml(&p)
			}
		}

		UpdateClientConfig()

		event.EntitiesUpdated("photos", approved)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionApproved))
	})
}

// BatchAlbumsDelete permanently removes multiple albums.
//
//	@Summary	permanently removes multiple albums
//	@Id			BatchAlbumsDelete
//	@Tags		Albums
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

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		// Get album UIDs.
		albumUIDs := f.Albums

		if len(albumUIDs) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoAlbumsSelected)
			return
		}

		log.Infof("albums: deleting %s", clean.Log(f.String()))

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
					SaveAlbumYaml(a)
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

// BatchPhotosPrivate toggles private state of multiple photos.
//
//	@Summary	toggles private state of multiple photos
//	@Id			BatchPhotosPrivate
//	@Tags		Photos
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

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if len(f.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		log.Infof("photos: updating private flag for %s", clean.Log(f.String()))

		if err := entity.Db().Model(entity.Photo{}).Where("photo_uid IN (?)", f.Photos).UpdateColumn("photo_private",
			gorm.Expr("CASE WHEN photo_private > 0 THEN 0 ELSE 1 END")).Error; err != nil {
			log.Errorf("private: %s", err)
			AbortSaveFailed(c)
			return
		}

		// Update precalculated photo and file counts.
		logWarn("index", entity.UpdateCounts())

		// Fetch selection from index.
		if photos, err := query.SelectedPhotos(f); err == nil {
			for _, p := range photos {
				SaveSidecarYaml(&p)
			}

			event.EntitiesUpdated("photos", photos)
		}

		UpdateClientConfig()

		FlushCoverCache()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionProtected))
	})
}

// BatchLabelsDelete deletes multiple labels.
//
//	@Summary	deletes multiple labels
//	@Id			BatchLabelsDelete
//	@Tags		Labels
//	@Produce	json
//	@Success	200					{object}	i18n.Response
//	@Failure	400,401,403,429,500	{object}	i18n.Response
//	@Param		labels				body		form.Selection	true	"Label Selection"
//	@Router		/api/v1/batch/labels/delete [post]
func BatchLabelsDelete(router *gin.RouterGroup) {
	router.POST("/batch/labels/delete", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if len(f.Labels) == 0 {
			log.Error("no labels selected")
			Abort(c, http.StatusBadRequest, i18n.ErrNoLabelsSelected)
			return
		}

		log.Infof("labels: deleting %s", clean.Log(f.String()))

		var labels entity.Labels

		if err := entity.Db().Where("label_uid IN (?)", f.Labels).Find(&labels).Error; err != nil {
			Error(c, http.StatusInternalServerError, err, i18n.ErrDeleteFailed)
			return
		}

		for _, label := range labels {
			logErr("labels", label.Delete())
		}

		UpdateClientConfig()

		event.EntitiesDeleted("labels", f.Labels)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgLabelsDeleted))
	})
}

// BatchPhotosDelete permanently removes multiple photos from the archive.
//
//	@Summary	permanently removes multiple or all photos from the archive
//	@Id			BatchPhotosDelete
//	@Tags		Photos
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

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		deleteStart := time.Now()

		var photos entity.Photos
		var err error

		// Abort if user wants to delete all but does not have sufficient privileges.
		if f.All && !acl.Rules.AllowAll(acl.ResourcePhotos, s.UserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}) {
			AbortForbidden(c)
			return
		}

		// Get selection or all archived photos if f.All is true.
		if len(f.Photos) == 0 && !f.All {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		} else if f.All {
			photos, err = query.ArchivedPhotos(1000000, 0)
		} else {
			photos, err = query.SelectedPhotos(f)
		}

		// Abort if the query failed or no photos were found.
		if err != nil {
			log.Errorf("archive: %s", err)
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		} else if len(photos) > 0 {
			log.Infof("archive: deleting %s", english.Plural(len(photos), "photo", "photos"))
		} else {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		var deleted entity.Photos

		var numFiles = 0

		// Delete photos.
		for _, p := range photos {
			// Report file deletion.
			event.AuditWarn([]string{ClientIP(c), s.UserName, "delete", path.Join(p.PhotoPath, p.PhotoName+"*")})

			// Remove all related files from storage.
			n, deleteErr := photoprism.DeletePhoto(&p, true, true)

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
			// Update precalculated photo and file counts.
			logWarn("index", entity.UpdateCounts())

			UpdateClientConfig()

			event.EntitiesDeleted("photos", deleted.UIDs())
		}

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPermanentlyDeleted))
	})
}
