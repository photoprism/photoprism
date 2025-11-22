package api

import (
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

var albumMutex = sync.Mutex{}

// SaveAlbumYaml saves the album metadata to a YAML backup file.
func SaveAlbumYaml(album *entity.Album) {
	if album == nil {
		log.Debugf("api: album is nil (update yaml)")
		return
	} else if !album.HasID() {
		log.Debugf("api: album has no ID (update yaml)")
		return
	}

	conf := get.Config()

	// Check if saving YAML backup files is enabled.
	if !conf.BackupAlbums() {
		return
	}

	// Write album metadata to YAML backup file.
	_ = album.SaveBackupYaml(conf.BackupAlbumsPath())
}

// DeleteAlbumYaml removes the YAML backup file for the provided album if it exists.
func DeleteAlbumYaml(album entity.Album) {
	conf := get.Config()

	// Nothing to remove when album backups are disabled.
	if !conf.BackupAlbums() {
		return
	}

	fileName, relName, err := album.YamlFileName(conf.BackupAlbumsPath())

	if err != nil {
		log.Warnf("album: %s (delete %s)", err, clean.Log(relName))
		return
	}

	if !fs.FileExists(fileName) {
		return
	}

	if rmErr := os.Remove(fileName); rmErr != nil {
		log.Errorf("album: %s (delete %s)", rmErr, clean.Log(relName))
	}
}

// GetAlbum returns album details as JSON.
//
//	@Summary	returns album details as JSON
//	@Id			GetAlbum
//	@Tags		Albums
//	@Produce	json
//	@Success	200				{object}	entity.Album
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"Album UID"
//	@Router		/api/v1/albums/{uid} [get]
func GetAlbum(router *gin.RouterGroup) {
	router.GET("/albums/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionView)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors can only access shared content.
		if (s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		// Other restricted users can only access their own or shared content.
		if s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) && album.CreatedBy != s.UserUID && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		c.JSON(http.StatusOK, album)
	})
}

// CreateAlbum creates a new album.
//
//	@Summary		creates a new album
//	@Description	Posting a title that matches a soft-deleted manual album restores it (including existing photo assignments). Use DELETE with `force=true` to purge an album before recreating it from scratch.
//	@Id				CreateAlbum
//	@Tags			Albums
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	entity.Album
//	@Success		201					{object}	entity.Album
//	@Failure		400,401,403,429,500	{object}	i18n.Response
//	@Param			album				body		form.Album	true	"properties of the album to be created (currently supports Title and Favorite)"
//	@Router			/api/v1/albums [post]
func CreateAlbum(router *gin.RouterGroup) {
	router.POST("/albums", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionCreate)

		if s.Abort(c) {
			return
		}

		var frm form.Album

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		conf := get.Config()

		album := entity.NewUserAlbum(frm.AlbumTitle, entity.AlbumManual, conf.Settings().Albums.Order.Album, s.UserUID)
		album.AlbumFavorite = frm.AlbumFavorite

		code := http.StatusOK

		// Existing album?
		if found := album.Find(); found == nil {
			// Not found, create new album.
			if err := album.Create(); err != nil {
				// Report unexpected error.
				log.Errorf("album: %s (create)", err)
				AbortUnexpectedError(c)
				return
			}
			code = http.StatusCreated
		} else {
			// Exists, restore if necessary.
			album = found
			if !album.Deleted() {
				c.JSON(http.StatusOK, album)
				return
			} else if err := album.Restore(); err != nil {
				// Report unexpected error.
				log.Errorf("album: %s (restore)", err)
				AbortUnexpectedError(c)
				return
			}
		}

		UpdateClientConfig()

		// Update album YAML backup.
		SaveAlbumYaml(album)

		// Add location header if newly created.
		if code == http.StatusCreated {
			header.SetLocation(c, c.FullPath(), album.AlbumUID)
		}

		// Return as JSON.
		c.JSON(code, album)
	})
}

// UpdateAlbum updates album metadata like title and description.
//
//	@Summary	updates album metadata like title and description
//	@Id			UpdateAlbum
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	entity.Album
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		uid						path		string		true	"Album UID"
//	@Param		album					body		form.Album	true	"properties to be updated"
//	@Router		/api/v1/albums/{uid} [put]
func UpdateAlbum(router *gin.RouterGroup) {
	router.PUT("/albums/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		frm, err := form.NewAlbum(album)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Assign and validate request form values.
		if err = c.BindJSON(frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		if err = album.SaveForm(frm); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Flush album cover cache.
		RemoveFromAlbumCoverCache(uid)

		// Update client.
		UpdateClientConfig()

		// Update album YAML backup.
		SaveAlbumYaml(&album)

		c.JSON(http.StatusOK, album)
	})
}

// DeleteAlbum deletes an existing album.
//
//	@Summary	deletes an existing album
//	@Id			DeleteAlbum
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"Album UID"
//	@Param		force				query		boolean	false	"Set to true to permanently delete a manual album instead of archiving it."
//	@Router		/api/v1/albums/{uid} [delete]
func DeleteAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		albumMutex.Lock()
		defer albumMutex.Unlock()

		forceDelete := false

		if forceParam := c.Query("force"); forceParam != "" {
			if parsed, parseErr := strconv.ParseBool(forceParam); parseErr == nil {
				forceDelete = parsed
			}
		}

		// Regular, manually created album?
		if album.IsDefault() && !forceDelete {
			// Soft delete manually created albums.
			err = album.Delete()

			// Also update album YAML backup.
			if err != nil {
				log.Errorf("album: %s (delete)", err)
				AbortDeleteFailed(c)
				return
			} else {
				SaveAlbumYaml(&album)
			}
		} else {
			// Permanently delete automatically created albums.
			err = album.DeletePermanently()

			// Also remove YAML backup file, if it exists.
			if err != nil {
				log.Errorf("album: %s (delete permanently)", err)
				AbortDeleteFailed(c)
				return
			} else {
				DeleteAlbumYaml(album)
			}
		}

		UpdateClientConfig()

		c.JSON(http.StatusOK, album)
	})
}

// LikeAlbum sets the favorite flag for an album.
//
//	@Summary	sets the favorite flag for an album
//	@Id			LikeAlbum
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"Album UID"
//	@Router		/api/v1/albums/{uid}/like [post]
func LikeAlbum(router *gin.RouterGroup) {
	router.POST("/albums/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		if err = album.Update("AlbumFavorite", true); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		UpdateClientConfig()

		PublishAlbumEvent(StatusUpdated, uid, c)

		// Update album YAML backup.
		SaveAlbumYaml(&album)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgChangesSaved))
	})
}

// DislikeAlbum removes the favorite flag from an album.
//
//	@Summary	removes the favorite flag from an album
//	@Id			DislikeAlbum
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"Album UID"
//	@Router		/api/v1/albums/{uid}/like [delete]
func DislikeAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		if err = album.Update("AlbumFavorite", false); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		UpdateClientConfig()

		PublishAlbumEvent(StatusUpdated, uid, c)

		// Update album YAML backup.
		SaveAlbumYaml(&album)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgChangesSaved))
	})
}

// CloneAlbums creates a new album containing pictures from other albums.
//
//	@Summary	creates a new album containing pictures from other albums
//	@Id			CloneAlbums
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	gin.H
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		albums				body		form.Selection	true	"Album Selection"
//	@Param		uid					path		string			true	"UID of the album to which the pictures are to be added"
//	@Router		/api/v1/albums/{uid}/clone [post]
func CloneAlbums(router *gin.RouterGroup) {
	router.POST("/albums/:uid/clone", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionCreate)

		if s.Abort(c) {
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		var frm form.Selection

		// Assign and validate request form values.
		if err = c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		var added []entity.PhotoAlbum

		for _, albumUid := range frm.Albums {
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

			added = append(added, album.AddPhotos(photos)...)
		}

		if len(added) > 0 {
			event.SuccessMsg(i18n.MsgSelectionAddedTo, clean.Log(album.Title()))

			PublishAlbumEvent(StatusUpdated, album.AlbumUID, c)

			// Update album YAML backup.
			SaveAlbumYaml(&album)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgAlbumCloned), "album": album, "added": added})
	})
}

// AddPhotosToAlbum adds photos to an album.
//
//	@Summary	adds photos to an album
//	@Id			AddPhotosToAlbum
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	gin.H
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		photos				body		form.Selection	true	"Photo Selection"
//	@Param		uid					path		string			true	"Album UID"
//	@Router		/api/v1/albums/{uid}/photos [post]
func AddPhotosToAlbum(router *gin.RouterGroup) {
	router.POST("/albums/:uid/photos", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		var frm form.Selection

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		} else if !album.HasID() {
			AbortAlbumNotFound(c)
			return
		} else if frm.Empty() {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		// Fetch selection from index.
		photos, err := query.SelectedPhotos(frm)

		if err != nil {
			AbortBadRequest(c, err)
			return
		}

		conf := get.Config()

		added := album.AddPhotos(photos)

		if len(added) > 0 {
			if len(added) == 1 {
				event.SuccessMsg(i18n.MsgEntryAddedTo, clean.Log(album.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesAddedTo, len(added), clean.Log(album.Title()))
			}

			RemoveFromAlbumCoverCache(album.AlbumUID)

			PublishAlbumEvent(StatusUpdated, album.AlbumUID, c)

			// Update album YAML backup.
			SaveAlbumYaml(&album)

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
						SaveSidecarYaml(p)
					}
				}

				// Update client UI and counts if photos has been approved.
				if len(approved) > 0 {
					UpdateClientConfig()

					event.EntitiesUpdated("photos", approved)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": album, "photos": photos.UIDs(), "added": added})
	})
}

// RemovePhotosFromAlbum removes photos from an album.
//
//	@Summary	removes photos from an album
//	@Id			RemovePhotosFromAlbum
//	@Tags		Albums
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	gin.H
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		photos				body		form.Selection	true	"Photo Selection"
//	@Param		uid					path		string			true	"Album UID"
//	@Router		/api/v1/albums/{uid}/photos [delete]
func RemovePhotosFromAlbum(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/photos", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionUpdate)

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

		// Get sanitized album UID from request path.
		uid := clean.UID(c.Param("uid"))

		// Visitors and other restricted users can only access shared content.
		if (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid) {
			AbortForbidden(c)
			return
		}

		// Find album by UID.
		album, err := query.AlbumByUID(uid)

		if err != nil {
			AbortAlbumNotFound(c)
			return
		} else if !album.HasID() {
			AbortAlbumNotFound(c)
			return
		}

		removed := album.RemovePhotos(frm.Photos)

		if len(removed) > 0 {
			if len(removed) == 1 {
				event.SuccessMsg(i18n.MsgEntryRemovedFrom, clean.Log(album.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesRemovedFrom, len(removed), clean.Log(album.Title()))
			}

			RemoveFromAlbumCoverCache(album.AlbumUID)

			PublishAlbumEvent(StatusUpdated, album.AlbumUID, c)

			// Update album YAML backup.
			SaveAlbumYaml(&album)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": album, "photos": frm.Photos, "removed": removed})
	})
}
