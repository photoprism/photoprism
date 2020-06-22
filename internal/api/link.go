package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// PUT /api/v1/:entity/:uid/links/:link
func UpdateLink(c *gin.Context, conf *config.Config) {
	if Unauthorized(c, conf) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var f form.Link

	if err := c.BindJSON(&f); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
		return
	}

	link := entity.FindLink(c.Param("link"))

	link.ShareExpires = f.ShareExpires

	if f.ShareToken != "" {
		link.ShareToken = strings.ToLower(f.ShareToken)
	}

	if f.Password != "" {
		if err := link.SetPassword(f.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}
	}

	if err := link.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UcFirst(err.Error())})
		return
	}

	event.Success("updated share link")

	c.JSON(http.StatusOK, link)
}

// DELETE /api/v1/:entity/:uid/links/:link
func DeleteLink(c *gin.Context, conf *config.Config) {
	if Unauthorized(c, conf) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	link := entity.FindLink(c.Param("link"))

	if err := link.Delete(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UcFirst(err.Error())})
		return
	}

	event.Success("deleted share link")

	c.JSON(http.StatusOK, link)
}

// CreateLink returns a new link entity initialized with request data
func CreateLink(c *gin.Context, conf *config.Config) {
	if Unauthorized(c, conf) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var f form.Link

	if err := c.BindJSON(&f); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
		return
	}

	link := entity.NewLink(c.Param("uid"), f.CanComment, f.CanEdit)

	if f.ShareExpires > 0 {
		link.ShareExpires = f.ShareExpires
	}

	if f.Password != "" {
		if err := link.SetPassword(f.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}
	}

	if err := link.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UcFirst(err.Error())})
		return
	}

	event.Success("added share link")

	c.JSON(http.StatusOK, link)
}

// POST /api/v1/albums/:uid/links
func CreateAlbumLink(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/albums/:uid/links", func(c *gin.Context) {
		if _, err := query.AlbumByUID(c.Param("uid")); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		CreateLink(c, conf)
	})
}

// PUT /api/v1/albums/:uid/links/:link
func UpdateAlbumLink(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/albums/:uid/links/:link", func(c *gin.Context) {
		UpdateLink(c, conf)
	})
}

// DELETE /api/v1/albums/:uid/links/:link
func DeleteAlbumLink(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/albums/:uid/links/:link", func(c *gin.Context) {
		DeleteLink(c, conf)
	})
}

// GET /api/v1/albums/:uid/links
func GetAlbumLinks(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uid/links", func(c *gin.Context) {
		m, err := query.AlbumByUID(c.Param("uid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}

// POST /api/v1/photos/:uid/links
func CreatePhotoLink(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/photos/:uid/links", func(c *gin.Context) {
		if _, err := query.PhotoByUID(c.Param("uid")); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		CreateLink(c, conf)
	})
}

// PUT /api/v1/photos/:uid/links/:link
func UpdatePhotoLink(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/photos/:uid/links/:link", func(c *gin.Context) {
		UpdateLink(c, conf)
	})
}

// DELETE /api/v1/photos/:uid/links/:link
func DeletePhotoLink(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/photos/:uid/links/:link", func(c *gin.Context) {
		DeleteLink(c, conf)
	})
}

// GET /api/v1/photos/:uid/links
func GetPhotoLinks(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/photos/:uid/links", func(c *gin.Context) {
		m, err := query.PhotoByUID(c.Param("uid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}

// POST /api/v1/labels/:uid/links
func CreateLabelLink(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/labels/:uid/links", func(c *gin.Context) {
		if _, err := query.LabelByUID(c.Param("uid")); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrLabelNotFound)
			return
		}

		CreateLink(c, conf)
	})
}


// PUT /api/v1/labels/:uid/links/:link
func UpdateLabelLink(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/labels/:uid/links/:link", func(c *gin.Context) {
		UpdateLink(c, conf)
	})
}

// DELETE /api/v1/labels/:uid/links/:link
func DeleteLabelLink(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/labels/:uid/links/:link", func(c *gin.Context) {
		DeleteLink(c, conf)
	})
}

// GET /api/v1/labels/:uid/links
func GetLabelLinks(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels/:uid/links", func(c *gin.Context) {
		m, err := query.LabelByUID(c.Param("uid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}
