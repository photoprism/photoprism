package api

import (
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

var albumMutex = sync.Mutex{}

// SaveAlbumYaml saves the album metadata to a YAML backup file.
func SaveAlbumYaml(album entity.Album) {
	conf := get.Config()

	// Check if saving YAML backup files is enabled.
	if !conf.BackupAlbums() {
		return
	}

	// Write album metadata to YAML backup file.
	_ = album.SaveBackupYaml(conf.BackupAlbumsPath())
}

// GetAlbum returns album details as JSON.
//
// GET /api/v1/albums/:uid
func GetAlbum(router *gin.RouterGroup) {
	router.GET("/albums/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionView)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
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
		s := Auth(c, acl.ResourceAlbums, acl.ActionCreate)

		if s.Abort(c) {
			return
		}

		var f form.Album

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		a := entity.NewUserAlbum(f.AlbumTitle, entity.AlbumManual, s.UserUID)
		a.AlbumFavorite = f.AlbumFavorite

		// Existing album?
		if found := a.Find(); found == nil {
			// Not found, create new album.
			if err := a.Create(); err != nil {
				// Report unexpected error.
				log.Errorf("album: %s (create)", err)
				AbortUnexpectedError(c)
				return
			}
		} else {
			// Exists, restore if necessary.
			a = found
			if !a.Deleted() {
				c.JSON(http.StatusOK, a)
				return
			} else if err := a.Restore(); err != nil {
				// Report unexpected error.
				log.Errorf("album: %s (restore)", err)
				AbortUnexpectedError(c)
				return
			}
		}

		UpdateClientConfig()

		// Update album YAML backup.
		SaveAlbumYaml(*a)

		// Return as JSON.
		c.JSON(http.StatusOK, a)
	})
}

// UpdateAlbum updates album metadata like title and description.
//
// PUT /api/v1/albums/:uid
func UpdateAlbum(router *gin.RouterGroup) {
	router.PUT("/albums/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		f, err := form.NewAlbum(a)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Assign and validate request form values.
		if err = c.BindJSON(&f); err != nil {
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

		// Flush album cover cache.
		RemoveFromAlbumCoverCache(uid)

		// Update client.
		UpdateClientConfig()

		// Update album YAML backup.
		SaveAlbumYaml(a)

		c.JSON(http.StatusOK, a)
	})
}

// DeleteAlbum deletes an existing album.
//
// DELETE /api/v1/albums/:uid
func DeleteAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		// Regular, manually created album?
		if a.IsDefault() {
			// Soft delete manually created albums.
			err = a.Delete()

			// Also update album YAML backup.
			if err != nil {
				log.Errorf("album: %s (delete)", err)
				AbortDeleteFailed(c)
				return
			} else {
				SaveAlbumYaml(a)
			}
		} else {
			// Permanently delete automatically created albums.
			err = a.DeletePermanently()

			// Also remove YAML backup file, if it exists.
			if err != nil {
				log.Errorf("album: %s (delete permanently)", err)
				AbortDeleteFailed(c)
				return
			} else if fileName, relName, nameErr := a.YamlFileName(get.Config().BackupAlbumsPath()); nameErr != nil {
				log.Warnf("album: %s (delete %s)", err, clean.Log(relName))
			} else if !fs.FileExists(fileName) {
				// Do nothing.
			} else if removeErr := os.Remove(fileName); removeErr != nil {
				log.Errorf("album: %s (delete %s)", err, clean.Log(relName))
			}
		}

		UpdateClientConfig()

		c.JSON(http.StatusOK, a)
	})
}

// LikeAlbum sets the favorite flag for an album.
//
// The request parameters are:
//
//   - uid: string Album UID
//
// POST /api/v1/albums/:uid/like
func LikeAlbum(router *gin.RouterGroup) {
	router.POST("/albums/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		if err := a.Update("AlbumFavorite", true); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		UpdateClientConfig()

		PublishAlbumEvent(StatusUpdated, uid, c)

		// Update album YAML backup.
		SaveAlbumYaml(a)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgChangesSaved))
	})
}

// DislikeAlbum removes the favorite flag from an album.
//
// The request parameters are:
//
//   - uid: string Album UID
//
// DELETE /api/v1/albums/:uid/like
func DislikeAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		if err = a.Update("AlbumFavorite", false); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		UpdateClientConfig()

		PublishAlbumEvent(StatusUpdated, uid, c)

		// Update album YAML backup.
		SaveAlbumYaml(a)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgChangesSaved))
	})
}

// CloneAlbums creates a new album containing pictures from other albums.
//
// POST /api/v1/albums/:uid/clone
func CloneAlbums(router *gin.RouterGroup) {
	router.POST("/albums/:uid/clone", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionCreate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		var f form.Selection

		// Assign and validate request form values.
		if err = c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		var added []entity.PhotoAlbum

		for _, albumUid := range f.Albums {
			cloneAlbum, queryErr := query.AlbumByUID(albumUid)

			if queryErr != nil {
				log.Errorf("album: %s", queryErr)
				continue
			}

			photos, queryErr := search.AlbumPhotos(cloneAlbum, 100000, false)

			if queryErr != nil {
				log.Errorf("album: %s", queryErr)
				continue
			}

			added = append(added, a.AddPhotos(photos)...)
		}

		if len(added) > 0 {
			event.SuccessMsg(i18n.MsgSelectionAddedTo, clean.Log(a.Title()))

			PublishAlbumEvent(StatusUpdated, a.AlbumUID, c)

			// Update album YAML backup.
			SaveAlbumYaml(a)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgAlbumCloned), "album": a, "added": added})
	})
}

// AddPhotosToAlbum adds photos to an album.
//
// POST /api/v1/albums/:uid/photos
func AddPhotosToAlbum(router *gin.RouterGroup) {
	router.POST("/albums/:uid/photos", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		var f form.Selection

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		} else if !a.HasID() {
			AbortAlbumNotFound(c)
			return
		} else if f.Empty() {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		// Fetch selection from index.
		photos, err := query.SelectedPhotos(f)

		if err != nil {
			log.Errorf("album: %s", err)
			AbortBadRequest(c)
			return
		}

		conf := get.Config()

		added := a.AddPhotos(photos)

		if len(added) > 0 {
			if len(added) == 1 {
				event.SuccessMsg(i18n.MsgEntryAddedTo, clean.Log(a.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesAddedTo, len(added), clean.Log(a.Title()))
			}

			RemoveFromAlbumCoverCache(a.AlbumUID)

			PublishAlbumEvent(StatusUpdated, a.AlbumUID, c)

			// Update album YAML backup.
			SaveAlbumYaml(a)

			// Auto-approve photos that have been added to an album,
			// see https://github.com/photoprism/photoprism/issues/4229
			if conf.Settings().Features.Review {
				var approved entity.Photos

				for _, p := range photos {
					// Skip photos that are not in review.
					if p.Approved() {
						continue
					}

					// Approve photo and update YAML backup file.
					if err = p.Approve(); err != nil {
						log.Errorf("approve: %s", err)
					} else {
						approved = append(approved, p)
						SaveSidecarYaml(&p)
					}
				}

				// Update client UI and counts if photos has been approved.
				if len(approved) > 0 {
					UpdateClientConfig()

					event.EntitiesUpdated("photos", approved)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": a, "photos": photos.UIDs(), "added": added})
	})
}

// RemovePhotosFromAlbum removes photos from an album.
//
// DELETE /api/v1/albums/:uid/photos
func RemovePhotosFromAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/photos", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

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

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.User().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		a, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		} else if !a.HasID() {
			AbortAlbumNotFound(c)
			return
		}

		removed := a.RemovePhotos(f.Photos)

		if len(removed) > 0 {
			if len(removed) == 1 {
				event.SuccessMsg(i18n.MsgEntryRemovedFrom, clean.Log(a.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesRemovedFrom, len(removed), clean.Log(a.Title()))
			}

			RemoveFromAlbumCoverCache(a.AlbumUID)

			PublishAlbumEvent(StatusUpdated, a.AlbumUID, c)

			// Update album YAML backup.
			SaveAlbumYaml(a)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": a, "photos": f.Photos, "removed": removed})
	})
}
