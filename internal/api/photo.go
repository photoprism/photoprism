package api

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SavePhotoAsYaml saves photo data as YAML file.
func SavePhotoAsYaml(p entity.Photo, conf *config.Config) {
	// Write YAML sidecar file (optional).
	if conf.SidecarYaml() {
		yamlFile := p.YamlFileName(conf.OriginalsPath(), conf.SidecarHidden())

		if err := p.SaveAsYaml(yamlFile); err != nil {
			log.Errorf("photo: %s (update yaml)", err)
		} else {
			log.Infof("photo: updated yaml file %s", txt.Quote(fs.RelativeName(yamlFile, conf.OriginalsPath())))
		}
	}
}

// GET /api/v1/photos/:uid
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func GetPhoto(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/photos/:uid", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		p, err := query.PhotoPreloadByUID(c.Param("uid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		c.IndentedJSON(http.StatusOK, p)
	})
}

// PUT /api/v1/photos/:uid
func UpdatePhoto(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/photos/:uid", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		uid := c.Param("uid")
		m, err := query.PhotoByUID(uid)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		// TODO: Proof-of-concept for form handling - might need refactoring
		// 1) Init form with model values
		f, err := form.NewPhoto(m)

		if err != nil {
			log.Errorf("photo: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		// 2) Update form with values from request
		if err := c.BindJSON(&f); err != nil {
			log.Errorf("photo: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrFormInvalid)
			return
		}

		// 3) Save model with values from form
		if err := entity.SavePhotoForm(m, f, conf.GeoCodingApi()); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		PublishPhotoEvent(EntityUpdated, uid, c)

		event.Success("photo saved")

		p, err := query.PhotoPreloadByUID(uid)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		SavePhotoAsYaml(p, conf)

		c.JSON(http.StatusOK, p)
	})
}

// GET /api/v1/photos/:uid/dl
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func GetPhotoDownload(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/photos/:uid/dl", func(c *gin.Context) {
		if InvalidDownloadToken(c, conf) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		f, err := query.FileByPhotoUID(c.Param("uid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("photo: file %s is missing", txt.Quote(f.FileName))
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("photo", f.Update("FileMissing", true))

			return
		}

		downloadFileName := f.ShareFileName()

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))

		c.File(fileName)
	})
}

// GET /api/v1/photos/:uid/yaml
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func GetPhotoYaml(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/photos/:uid/yaml", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		p, err := query.PhotoPreloadByUID(c.Param("uid"))

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
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", c.Param("uid")+fs.YamlExt))
		}

		c.Data(http.StatusOK, "text/x-yaml; charset=utf-8", data)
	})
}

// POST /api/v1/photos/:uid/like
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func LikePhoto(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/photos/:uid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := c.Param("uid")
		m, err := query.PhotoByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		if err := m.SetFavorite(true); err != nil {
			log.Errorf("photo: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		SavePhotoAsYaml(m, conf)

		PublishPhotoEvent(EntityUpdated, id, c)

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// DELETE /api/v1/photos/:uid/like
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func DislikePhoto(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/photos/:uid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := c.Param("uid")
		m, err := query.PhotoByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		if err := m.SetFavorite(false); err != nil {
			log.Errorf("photo: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		SavePhotoAsYaml(m, conf)

		PublishPhotoEvent(EntityUpdated, id, c)

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// POST /api/v1/photos/:uid/primary/:file_uid
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func SetPhotoPrimary(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/photos/:uid/primary/:file_uid", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		uid := c.Param("uid")
		fileUID := c.Param("file_uid")
		err := query.SetPhotoPrimary(uid, fileUID)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		PublishPhotoEvent(EntityUpdated, uid, c)

		event.Success("photo saved")

		p, err := query.PhotoPreloadByUID(uid)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
