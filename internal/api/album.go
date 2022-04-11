package api

import (
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

var albumMutex = sync.Mutex{}

// SaveAlbumAsYaml saves album data as YAML file.
func SaveAlbumAsYaml(a entity.Album) {
	c := service.Config()

	// Write YAML sidecar file (optional).
	if !c.BackupYaml() {
		return
	}

	fileName := a.YamlFileName(c.AlbumsPath())

	if err := a.SaveAsYaml(fileName); err != nil {
		log.Errorf("album: %s (update yaml)", err)
	} else {
		log.Debugf("album: updated yaml file %s", sanitize.Log(filepath.Base(fileName)))
	}
}

// GetAlbum returns album details as JSON.
//
// GET /api/v1/albums/:uid
func GetAlbum(router *gin.RouterGroup) {
	router.GET("/albums/:uid", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdString(c.Param("uid"))
		a, err := query.AlbumByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		c.JSON(http.StatusOK, a)
	})
}

// CreateAlbum adds a new album.
//
// POST /api/v1/albums
func CreateAlbum(router *gin.RouterGroup) {
	router.POST("/albums", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionCreate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.Album

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		a := entity.NewAlbum(f.AlbumTitle, entity.AlbumDefault)
		a.AlbumFavorite = f.AlbumFavorite

		// Search existing album.
		if err := a.Find(); err == nil {
			c.JSON(http.StatusOK, a)
			return
		}

		// Create new album.
		if err := a.Create(); err != nil {
			AbortAlreadyExists(c, sanitize.Log(a.AlbumTitle))
			return
		}

		event.SuccessMsg(i18n.MsgAlbumCreated)

		UpdateClientConfig()

		PublishAlbumEvent(EntityCreated, a.AlbumUID, c)

		SaveAlbumAsYaml(*a)

		c.JSON(http.StatusOK, a)
	})
}

// UpdateAlbum updates album metadata like title and description.
//
// PUT /api/v1/albums/:uid
func UpdateAlbum(router *gin.RouterGroup) {
	router.PUT("/albums/:uid", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		uid := sanitize.IdString(c.Param("uid"))
		a, err := query.AlbumByUID(uid)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		f, err := form.NewAlbum(a)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		if err := c.BindJSON(&f); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		if err = a.SaveForm(f); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		UpdateClientConfig()

		event.SuccessMsg(i18n.MsgAlbumSaved)

		PublishAlbumEvent(EntityUpdated, uid, c)

		SaveAlbumAsYaml(a)

		c.JSON(http.StatusOK, a)
	})
}

// DeleteAlbum deletes an existing album.
//
// DELETE /api/v1/albums/:uid
func DeleteAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdString(c.Param("uid"))

		a, err := query.AlbumByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		// Regular, manually created album?
		if a.IsDefault() {
			// Soft delete manually created albums.
			err = a.Delete()
		} else {
			// Permanently delete automatically created albums.
			err = a.DeletePermanently()
		}

		if err != nil {
			log.Errorf("album: %s (delete)", err)
			AbortDeleteFailed(c)
			return
		}

		PublishAlbumEvent(EntityDeleted, id, c)

		UpdateClientConfig()

		SaveAlbumAsYaml(a)

		event.SuccessMsg(i18n.MsgAlbumDeleted, sanitize.Log(a.AlbumTitle))

		c.JSON(http.StatusOK, a)
	})
}

// LikeAlbum sets the favorite flag for an album.
//
// POST /api/v1/albums/:uid/like
//
// Parameters:
//   uid: string Album UID
func LikeAlbum(router *gin.RouterGroup) {
	router.POST("/albums/:uid/like", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionLike)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdString(c.Param("uid"))
		a, err := query.AlbumByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		if err := a.Update("AlbumFavorite", true); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		UpdateClientConfig()

		PublishAlbumEvent(EntityUpdated, id, c)

		SaveAlbumAsYaml(a)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgChangesSaved))
	})
}

// DislikeAlbum removes the favorite flag from an album.
//
// DELETE /api/v1/albums/:uid/like
//
// Parameters:
//   uid: string Album UID
func DislikeAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/like", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionLike)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdString(c.Param("uid"))
		a, err := query.AlbumByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		if err := a.Update("AlbumFavorite", false); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		UpdateClientConfig()

		PublishAlbumEvent(EntityUpdated, id, c)

		SaveAlbumAsYaml(a)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgChangesSaved))
	})
}

// CloneAlbums creates a new album containing pictures from other albums.
//
// POST /api/v1/albums/:uid/clone
func CloneAlbums(router *gin.RouterGroup) {
	router.POST("/albums/:uid/clone", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		a, err := query.AlbumByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		var added []entity.PhotoAlbum

		for _, uid := range f.Albums {
			cloneAlbum, err := query.AlbumByUID(uid)

			if err != nil {
				log.Errorf("album: %s", err)
				continue
			}

			photos, err := search.AlbumPhotos(cloneAlbum, 10000, false)

			if err != nil {
				log.Errorf("album: %s", err)
				continue
			}

			added = append(added, a.AddPhotos(photos.UIDs())...)
		}

		if len(added) > 0 {
			event.SuccessMsg(i18n.MsgSelectionAddedTo, sanitize.Log(a.Title()))

			PublishAlbumEvent(EntityUpdated, a.AlbumUID, c)

			SaveAlbumAsYaml(a)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgAlbumCloned), "album": a, "added": added})
	})
}

// AddPhotosToAlbum adds photos to an album.
//
// POST /api/v1/albums/:uid/photos
func AddPhotosToAlbum(router *gin.RouterGroup) {
	router.POST("/albums/:uid/photos", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		uid := sanitize.IdString(c.Param("uid"))
		a, err := query.AlbumByUID(uid)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		// Fetch selection from index.
		photos, err := query.SelectedPhotos(f)

		if err != nil {
			log.Errorf("album: %s", err)
			AbortBadRequest(c)
			return
		}

		added := a.AddPhotos(photos.UIDs())

		if len(added) > 0 {
			if len(added) == 1 {
				event.SuccessMsg(i18n.MsgEntryAddedTo, sanitize.Log(a.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesAddedTo, len(added), sanitize.Log(a.Title()))
			}

			RemoveFromAlbumCoverCache(a.AlbumUID)

			PublishAlbumEvent(EntityUpdated, a.AlbumUID, c)

			SaveAlbumAsYaml(a)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": a, "photos": photos.UIDs(), "added": added})
	})
}

// RemovePhotosFromAlbum removes photos from an album.
//
// DELETE /api/v1/albums/:uid/photos
func RemovePhotosFromAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/photos", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionUpdate)

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

		a, err := query.AlbumByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		removed := a.RemovePhotos(f.Photos)

		if len(removed) > 0 {
			if len(removed) == 1 {
				event.SuccessMsg(i18n.MsgEntryRemovedFrom, sanitize.Log(a.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesRemovedFrom, len(removed), sanitize.Log(sanitize.Log(a.Title())))
			}

			RemoveFromAlbumCoverCache(a.AlbumUID)

			PublishAlbumEvent(EntityUpdated, a.AlbumUID, c)

			SaveAlbumAsYaml(a)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": a, "photos": f.Photos, "removed": removed})
	})
}
