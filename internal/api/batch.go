package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// BatchPhotosArchive moves multiple photos to the archive.
//
// POST /api/v1/batch/photos/archive
func BatchPhotosArchive(router *gin.RouterGroup) {
	router.POST("/batch/photos/archive", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
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

		log.Infof("photos: archiving %s", clean.Log(f.String()))

		if service.Config().BackupYaml() {
			// Fetch selection from index.
			photos, err := query.SelectedPhotos(f)

			if err != nil {
				AbortEntityNotFound(c)
				return
			}

			for _, p := range photos {
				if err := p.Archive(); err != nil {
					log.Errorf("archive: %s", err)
				} else {
					SavePhotoAsYaml(p)
				}
			}
		} else if err := entity.Db().Where("photo_uid IN (?)", f.Photos).Delete(&entity.Photo{}).Error; err != nil {
			log.Errorf("archive: %s", err)
			AbortSaveFailed(c)
			return
		} else if err := entity.Db().Model(&entity.PhotoAlbum{}).Where("photo_uid IN (?)", f.Photos).UpdateColumn("hidden", true).Error; err != nil {
			log.Errorf("archive: %s", err)
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
// POST /api/v1/batch/photos/restore
func BatchPhotosRestore(router *gin.RouterGroup) {
	router.POST("/batch/photos/restore", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
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

		if service.Config().BackupYaml() {
			// Fetch selection from index.
			photos, err := query.SelectedPhotos(f)

			if err != nil {
				AbortEntityNotFound(c)
				return
			}

			for _, p := range photos {
				if err := p.Restore(); err != nil {
					log.Errorf("restore: %s", err)
				} else {
					SavePhotoAsYaml(p)
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
// POST /api/v1/batch/photos/approve
func BatchPhotosApprove(router *gin.RouterGroup) {
	router.POST("batch/photos/approve", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
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
			if err := p.Approve(); err != nil {
				log.Errorf("approve: %s", err)
			} else {
				approved = append(approved, p)
				SavePhotoAsYaml(p)
			}
		}

		UpdateClientConfig()

		event.EntitiesUpdated("photos", approved)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionApproved))
	})
}

// BatchAlbumsDelete permanently removes multiple albums.
//
// POST /api/v1/batch/albums/delete
func BatchAlbumsDelete(router *gin.RouterGroup) {
	router.POST("/batch/albums/delete", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if len(f.Albums) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoAlbumsSelected)
			return
		}

		log.Infof("albums: deleting %s", clean.Log(f.String()))

		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.Album{})
		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.PhotoAlbum{})

		UpdateClientConfig()

		event.EntitiesDeleted("albums", f.Albums)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgAlbumsDeleted))
	})
}

// BatchPhotosPrivate flags multiple photos as private.
//
// POST /api/v1/batch/photos/private
func BatchPhotosPrivate(router *gin.RouterGroup) {
	router.POST("/batch/photos/private", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionPrivate)

		if s.Invalid() {
			AbortUnauthorized(c)
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
				SavePhotoAsYaml(p)
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
// POST /api/v1/batch/labels/delete
func BatchLabelsDelete(router *gin.RouterGroup) {
	router.POST("/batch/labels/delete", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceLabels, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
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
			logError("labels", label.Delete())
		}

		UpdateClientConfig()

		event.EntitiesDeleted("labels", f.Labels)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgLabelsDeleted))
	})
}

// BatchPhotosDelete permanently removes multiple photos from the archive.
//
// POST /api/v1/batch/photos/delete
func BatchPhotosDelete(router *gin.RouterGroup) {
	router.POST("/batch/photos/delete", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Delete {
			AbortFeatureDisabled(c)
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

		log.Infof("photos: deleting %s", clean.Log(f.String()))

		// Fetch selection from index.
		photos, err := query.SelectedPhotos(f)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var deleted entity.Photos

		// Delete photos.
		for _, p := range photos {
			if err := photoprism.Delete(p); err != nil {
				log.Errorf("delete: %s", err)
			} else {
				deleted = append(deleted, p)
			}
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
