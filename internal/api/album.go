package api

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

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
		log.Debugf("album: updated yaml file %s", txt.Quote(filepath.Base(fileName)))
	}
}

// SearchAlbums finds albums and returns them as JSON.
//
// GET /api/v1/albums
func SearchAlbums(router *gin.RouterGroup) {
	router.GET("/albums", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.AlbumSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// Guest permissions are limited to shared albums.
		if s.Guest() {
			f.ID = s.Shares.Join(txt.Or)
		}

		result, err := search.Albums(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		c.JSON(http.StatusOK, result)
	})
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

		id := c.Param("uid")
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

		a := entity.NewAlbum(f.AlbumTitle, entity.AlbumDefault)
		a.AlbumFavorite = f.AlbumFavorite

		log.Debugf("album: creating %+v %+v", f, a)

		if res := entity.Db().Create(a); res.Error != nil {
			AbortAlreadyExists(c, txt.Quote(a.AlbumTitle))
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

		uid := c.Param("uid")
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

		if err := a.SaveForm(f); err != nil {
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

		id := c.Param("uid")

		a, err := query.AlbumByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

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

		event.SuccessMsg(i18n.MsgAlbumDeleted, txt.Quote(a.AlbumTitle))

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

		id := c.Param("uid")
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

		id := c.Param("uid")
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

		a, err := query.AlbumByUID(c.Param("uid"))

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

			photos, err := search.AlbumPhotos(cloneAlbum, 10000)

			if err != nil {
				log.Errorf("album: %s", err)
				continue
			}

			added = append(added, a.AddPhotos(photos.UIDs())...)
		}

		if len(added) > 0 {
			event.SuccessMsg(i18n.MsgSelectionAddedTo, txt.Quote(a.Title()))

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

		uid := c.Param("uid")
		a, err := query.AlbumByUID(uid)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		photos, err := query.PhotoSelection(f)

		if err != nil {
			log.Errorf("album: %s", err)
			AbortBadRequest(c)
			return
		}

		added := a.AddPhotos(photos.UIDs())

		if len(added) > 0 {
			if len(added) == 1 {
				event.SuccessMsg(i18n.MsgEntryAddedTo, txt.Quote(a.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesAddedTo, len(added), txt.Quote(a.Title()))
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

		a, err := query.AlbumByUID(c.Param("uid"))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		removed := a.RemovePhotos(f.Photos)

		if len(removed) > 0 {
			if len(removed) == 1 {
				event.SuccessMsg(i18n.MsgEntryRemovedFrom, txt.Quote(a.Title()))
			} else {
				event.SuccessMsg(i18n.MsgEntriesRemovedFrom, len(removed), txt.Quote(txt.Quote(a.Title())))
			}

			RemoveFromAlbumCoverCache(a.AlbumUID)

			PublishAlbumEvent(EntityUpdated, a.AlbumUID, c)

			SaveAlbumAsYaml(a)
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgChangesSaved), "album": a, "photos": f.Photos, "removed": removed})
	})
}

// DownloadAlbum streams the album contents as zip archive.
//
// GET /api/v1/albums/:uid/dl
func DownloadAlbum(router *gin.RouterGroup) {
	router.GET("/albums/:uid/dl", func(c *gin.Context) {
		if InvalidDownloadToken(c) {
			AbortUnauthorized(c)
			return
		}

		start := time.Now()
		a, err := query.AlbumByUID(c.Param("uid"))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
			return
		}

		files, err := search.AlbumPhotos(a, 10000)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		zipFileName := a.ZipName()

		AddDownloadHeader(c, zipFileName)

		zipWriter := zip.NewWriter(c.Writer)
		defer func() { _ = zipWriter.Close() }()

		var aliases = make(map[string]int)

		for _, file := range files {
			if file.FileHash == "" {
				log.Warnf("download: empty file hash, skipped %s", txt.Quote(file.FileName))
				continue
			}

			if file.FileSidecar {
				log.Debugf("download: skipped sidecar %s", txt.Quote(file.FileName))
				continue
			}

			fileName := photoprism.FileName(file.FileRoot, file.FileName)
			alias := file.ShareBase(0)
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.ShareBase(seq)
			}

			aliases[key] += 1

			if fs.FileExists(fileName) {
				if err := addFileToZip(zipWriter, fileName, alias); err != nil {
					log.Error(err)
					Abort(c, http.StatusInternalServerError, i18n.ErrZipFailed)
					return
				}
				log.Infof("download: added %s as %s", txt.Quote(file.FileName), txt.Quote(alias))
			} else {
				log.Errorf("download: failed finding %s", txt.Quote(file.FileName))
			}
		}

		log.Infof("download: created %s [%s]", txt.Quote(zipFileName), time.Since(start))
	})
}
