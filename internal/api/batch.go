package api

import (
	"net/http"
	"path"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// BatchPhotosArchive moves multiple photos to the archive.
//
// POST /api/v1/batch/photos/archive
func BatchPhotosArchive(router *gin.RouterGroup) {
	router.POST("/batch/photos/archive", func(c *gin.Context) {
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

		log.Infof("photos: archiving %s", clean.Log(f.String()))

		if get.Config().BackupYaml() {
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

		if get.Config().BackupYaml() {
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
		s := Auth(c, acl.ResourceAlbums, acl.ActionDelete)

		if s.Abort(c) {
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

		// Soft delete albums, can be restored.
		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.Album{})

		/*
			KEEP ENTRIES AS ALBUMS MAY NOW BE RESTORED BY NAME
			entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.PhotoAlbum{})
		*/

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
		if f.All && !acl.Resources.AllowAll(acl.ResourcePhotos, s.User().AclRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}) {
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
			n, err := photoprism.DeletePhoto(p, true, true)

			numFiles += n

			if err != nil {
				log.Errorf("delete: %s", err)
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
