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
	"github.com/photoprism/photoprism/internal/query"
)

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

		log.Infof("archive: adding %s", f.String())

		// Soft delete by setting deleted_at to current date.
		err := entity.Db().Where("photo_uid IN (?)", f.Photos).Delete(&entity.Photo{}).Error

		if err != nil {
			AbortSaveFailed(c)
			return
		}

		// Remove archived photos from albums.
		logError("archive", entity.Db().Model(&entity.PhotoAlbum{}).Where("photo_uid IN (?)", f.Photos).UpdateColumn("hidden", true).Error)

		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("photos: %s", err)
		}

		UpdateClientConfig()

		event.EntitiesArchived("photos", f.Photos)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionArchived))
	})
}

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

		log.Infof("archive: restoring %s", f.String())

		err := entity.Db().Unscoped().Model(&entity.Photo{}).Where("photo_uid IN (?)", f.Photos).
			UpdateColumn("deleted_at", gorm.Expr("NULL")).Error

		if err != nil {
			AbortSaveFailed(c)
			return
		}

		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("photos: %s", err)
		}

		UpdateClientConfig()

		event.EntitiesRestored("photos", f.Photos)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionRestored))
	})
}

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

		log.Infof("albums: deleting %s", f.String())

		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.Album{})
		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.PhotoAlbum{})

		UpdateClientConfig()

		event.EntitiesDeleted("albums", f.Albums)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgAlbumsDeleted))
	})
}

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

		log.Infof("photos: mark %s as private", f.String())

		err := entity.Db().Model(entity.Photo{}).Where("photo_uid IN (?)", f.Photos).UpdateColumn("photo_private",
			gorm.Expr("CASE WHEN photo_private > 0 THEN 0 ELSE 1 END")).Error

		if err != nil {
			AbortSaveFailed(c)
			return
		}

		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("photos: %s", err)
		}

		if entities, err := query.PhotoSelection(f); err == nil {
			event.EntitiesUpdated("photos", entities)
		}

		UpdateClientConfig()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSelectionProtected))
	})
}

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

		log.Infof("labels: deleting %s", f.String())

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
