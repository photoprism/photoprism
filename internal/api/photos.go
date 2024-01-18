package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// SavePhotoAsYaml saves photo data as YAML file.
func SavePhotoAsYaml(p entity.Photo) {
	c := get.Config()

	// Write YAML sidecar file (optional).
	if !c.BackupYaml() {
		return
	}

	fileName := p.YamlFileName(c.OriginalsPath(), c.SidecarPath())

	if err := p.SaveAsYaml(fileName); err != nil {
		log.Errorf("photo: %s (update yaml)", err)
	} else {
		log.Debugf("photo: updated yaml file %s", clean.Log(filepath.Base(fileName)))
	}
}

// GetPhoto returns photo details as JSON.
//
// The request parameters are:
//
//   - uid (string) PhotoUID as returned by the API
//
// GET /api/v1/photos/:uid
func GetPhoto(router *gin.RouterGroup) {
	router.GET("/photos/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionView)

		if s.Abort(c) {
			return
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.IndentedJSON(http.StatusOK, p)
	})
}

// UpdatePhoto updates photo details and returns them as JSON.
//
// PUT /api/v1/photos/:uid
func UpdatePhoto(router *gin.RouterGroup) {
	router.PUT("/photos/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// 1) Init form with model values
		f, err := form.NewPhoto(m)

		if err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		}

		// 2) Update form with values from request
		if err := c.BindJSON(&f); err != nil {
			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		// 3) Save model with values from form
		if err := entity.SavePhotoForm(m, f); err != nil {
			Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
			return
		} else if f.PhotoPrivate {
			FlushCoverCache()
		}

		PublishPhotoEvent(StatusUpdated, uid, c)

		event.SuccessMsg(i18n.MsgChangesSaved)

		p, err := query.PhotoPreloadByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		SavePhotoAsYaml(p)

		UpdateClientConfig()

		c.JSON(http.StatusOK, p)
	})
}

// GetPhotoDownload returns the primary file matching that belongs to the photo.
//
// Route :GET /api/v1/photos/:uid/dl
//
// The request parameters are:
//
//   - uid (string) PhotoUID as returned by the API
func GetPhotoDownload(router *gin.RouterGroup) {
	router.GET("/photos/:uid/dl", func(c *gin.Context) {
		if InvalidDownloadToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		f, err := query.FileByPhotoUID(clean.UID(c.Param("uid")))

		if err != nil {
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("photo: file %s is missing", clean.Log(f.FileName))
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logErr("photo", f.Update("FileMissing", true))

			return
		}

		c.FileAttachment(fileName, f.DownloadName(DownloadName(c), 0))
	})
}

// GetPhotoYaml returns photo details as YAML.
//
// The request parameters are:
//
//   - uid: string PhotoUID as returned by the API
//
// GET /api/v1/photos/:uid/yaml
func GetPhotoYaml(router *gin.RouterGroup) {
	router.GET("/photos/:uid/yaml", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.AccessAll)

		if s.Abort(c) {
			return
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		data, err := p.Yaml()

		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if c.Query("download") != "" {
			AddDownloadHeader(c, clean.UID(c.Param("uid"))+fs.ExtYAML)
		}

		c.Data(http.StatusOK, "text/x-yaml; charset=utf-8", data)
	})
}

// ApprovePhoto marks a photo in review as approved.
//
// The request parameters are:
//
//   - uid: string PhotoUID as returned by the API
//
// POST /api/v1/photos/:uid/approve
func ApprovePhoto(router *gin.RouterGroup) {
	router.POST("/photos/:uid/approve", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := m.Approve(); err != nil {
			log.Errorf("photo: %s", err.Error())
			AbortSaveFailed(c)
			return
		}

		SavePhotoAsYaml(m)

		PublishPhotoEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// PhotoPrimary sets the primary file for a photo.
//
// The request parameters are:
//
//   - uid: string PhotoUID as returned by the API
//   - file_uid: string File UID as returned by the API
//
// POST /photos/:uid/files/:file_uid/primary
func PhotoPrimary(router *gin.RouterGroup) {
	router.POST("/photos/:uid/files/:file_uid/primary", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		fileUid := clean.UID(c.Param("file_uid"))
		err := query.SetPhotoPrimary(uid, fileUid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		PublishPhotoEvent(StatusUpdated, uid, c)

		event.SuccessMsg(i18n.MsgChangesSaved)

		p, err := query.PhotoPreloadByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
