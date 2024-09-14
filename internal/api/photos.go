package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// SaveSidecarYaml saves the photo metadata to a YAML sidecar file.
func SaveSidecarYaml(photo *entity.Photo) {
	if photo == nil {
		log.Debugf("api: photo is nil (update yaml)")
		return
	}

	conf := get.Config()

	// Check if saving YAML sidecar files is enabled.
	if !conf.SidecarYaml() {
		return
	}

	// Write photo metadata to YAML sidecar file.
	_ = photo.SaveSidecarYaml(conf.OriginalsPath(), conf.SidecarPath())
}

// GetPhoto returns picture details as JSON.
//
//	@Summary	returns picture details as JSON
//	@Id			GetPhoto
//	@Tags		Photos
//	@Produce	json
//	@Success	200				{object}	entity.Photo
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"Photo UID"
//	@Router		/api/v1/photos/{uid} [get]
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

// UpdatePhoto updates picture details and returns them as JSON.
//
//	@Summary	updates picture details and returns them as JSON
//	@Id			UpdatePhoto
//	@Tags		Photos
//	@Produce	json
//	@Success	200						{object}	entity.Photo
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		uid						path		string		true	"Photo UID"
//	@Param		photo					body		form.Photo	true	"properties to be updated (only submit values that should be changed)"
//	@Router		/api/v1/photos/{uid} [put]
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

		// 2) Assign and validate request form values.
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

		SaveSidecarYaml(&p)

		UpdateClientConfig()

		c.JSON(http.StatusOK, p)
	})
}

// GetPhotoDownload returns the primary file matching that belongs to the photo.
//
//	@Summary	returns the primary file matching that belongs to the photo
//	@Id			GetPhotoDownload
//	@Tags		Images, Files
//	@Produce	application/octet-stream
//	@Failure	403,404	{file}	image/svg+xml
//	@Success	200		{file}	application/octet-stream
//	@Param		uid		path	string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/dl [get]
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

// GetPhotoYaml returns picture details as YAML.
//
//	@Summary	returns picture details as YAML
//	@Id			GetPhotoYaml
//	@Tags		Photos
//	@Produce	text/x-yaml
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Success	200					{file}		text/x-yaml
//	@Param		uid					path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/yaml [get]
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
//	@Summary	marks a photo in review as approved
//	@Id			ApprovePhoto
//	@Tags		Photos
//	@Produce	json
//	@Success	200					{object}	gin.H
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/approve [post]
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

		SaveSidecarYaml(&m)

		PublishPhotoEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// PhotoPrimary sets the primary file for a photo.
//
//	@Summary	sets the primary file for a photo
//	@Id			PhotoPrimary
//	@Tags		Photos, Stacks
//	@Produce	json
//	@Success	200					{object}	entity.Photo
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"photo uid"
//	@Param		fileuid				path		string	true	"file uid"
//	@Router		/api/v1/photos/{uid}/files/{fileuid}/primary [post]
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
