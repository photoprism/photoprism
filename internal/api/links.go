package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UpdateLink updates a share link and return it as JSON.
//
// PUT /api/v1/:entity/:uid/links/:link
func UpdateLink(c *gin.Context) {
	s := Auth(c, acl.ResourceShares, acl.ActionUpdate)

	if s.Invalid() {
		AbortForbidden(c)
		return
	}

	var f form.Link

	// Assign and validate request form values.
	if err := c.BindJSON(&f); err != nil {
		log.Debugf("share: %s", err)
		AbortBadRequest(c)
		return
	}

	link := entity.FindLink(clean.Token(c.Param("link")))

	link.SetSlug(f.ShareSlug)
	link.MaxViews = f.MaxViews
	link.LinkExpires = f.LinkExpires

	if f.LinkToken != "" {
		link.LinkToken = strings.TrimSpace(strings.ToLower(f.LinkToken))
	}

	if f.Password != "" {
		if err := link.SetPassword(f.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}
	}

	if err := link.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
		return
	}

	UpdateClientConfig()

	PublishAlbumEvent(StatusUpdated, link.ShareUID, c)

	c.JSON(http.StatusOK, link)
}

// DeleteLink deletes a share link.
//
// DELETE /api/v1/:entity/:uid/links/:link
func DeleteLink(c *gin.Context) {
	s := Auth(c, acl.ResourceShares, acl.ActionDelete)

	if s.Invalid() {
		AbortForbidden(c)
		return
	}

	link := entity.FindLink(clean.Token(c.Param("link")))

	if err := link.Delete(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
		return
	}

	UpdateClientConfig()

	PublishAlbumEvent(StatusUpdated, link.ShareUID, c)

	c.JSON(http.StatusOK, link)
}

// CreateLink adds a new share link and return it as JSON.
//
// POST /api/v1/:entity/:uid/links
func CreateLink(c *gin.Context) {
	s := Auth(c, acl.ResourceShares, acl.ActionCreate)

	if s.Abort(c) {
		return
	}

	uid := clean.UID(c.Param("uid"))

	if uid == "" {
		AbortBadRequest(c)
		return
	}

	var f form.Link

	if err := c.BindJSON(&f); err != nil {
		log.Debugf("share: %s", err)
		AbortBadRequest(c)
		return
	}

	link := entity.NewUserLink(uid, s.UserUID)

	link.SetSlug(f.ShareSlug)
	link.MaxViews = f.MaxViews
	link.LinkExpires = f.LinkExpires

	if f.Password != "" {
		if err := link.SetPassword(f.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}
	}

	if err := link.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
		return
	}

	UpdateClientConfig()

	PublishAlbumEvent(StatusUpdated, link.ShareUID, c)

	c.JSON(http.StatusOK, link)
}

// CreateAlbumLink adds a new album share link and return it as JSON.
//
// POST /api/v1/albums/:uid/links
func CreateAlbumLink(router *gin.RouterGroup) {
	router.POST("/albums/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		if _, err := query.AlbumByUID(clean.UID(c.Param("uid"))); err != nil {
			AbortAlbumNotFound(c)
			return
		}

		CreateLink(c)
	})
}

// UpdateAlbumLink updates an album share link and return it as JSON.
//
// PUT /api/v1/albums/:uid/links/:link
func UpdateAlbumLink(router *gin.RouterGroup) {
	router.PUT("/albums/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		UpdateLink(c)
	})
}

// DeleteAlbumLink deletes an album share link.
//
// DELETE /api/v1/albums/:uid/links/:link
func DeleteAlbumLink(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		DeleteLink(c)
	})
}

// GetAlbumLinks returns all share links for the given UID as JSON.
//
// GET /api/v1/albums/:uid/links
func GetAlbumLinks(router *gin.RouterGroup) {
	router.GET("/albums/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		m, err := query.AlbumByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}

/*

// CreatePhotoLink adds a new photo share link and return it as JSON.
//
// POST /api/v1/photos/:uid/links
func CreatePhotoLink(router *gin.RouterGroup) {
	router.POST("/photos/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		if _, err := query.PhotoByUID(clean.UID(c.Param("uid"))); err != nil {
			AbortEntityNotFound(c)
			return
		}

		CreateLink(c)
	})
}

// UpdatePhotoLink updates an existing photo sharing link.
//
// PUT /api/v1/photos/:uid/links/:link
func UpdatePhotoLink(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		UpdateLink(c)
	})
}

// DeletePhotoLink deletes a photo sharing link.
//
// DELETE /api/v1/photos/:uid/links/:link
func DeletePhotoLink(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		DeleteLink(c)
	})
}

// GetPhotoLinks returns all share links for the given UID as JSON.
//
// GET /api/v1/photos/:uid/links
func GetPhotoLinks(router *gin.RouterGroup) {
	router.GET("/photos/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		m, err := query.PhotoByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}

// CreateLabelLink adds a new label share link and return it as JSON.
//
// POST /api/v1/labels/:uid/links
func CreateLabelLink(router *gin.RouterGroup) {
	router.POST("/labels/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		if _, err := query.LabelByUID(clean.UID(c.Param("uid"))); err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		CreateLink(c)
	})
}

// UpdateLabelLink updates a label share link and return it as JSON.
//
// PUT /api/v1/labels/:uid/links/:link
func UpdateLabelLink(router *gin.RouterGroup) {
	router.PUT("/labels/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		UpdateLink(c)
	})
}

// DeleteLabelLink deletes a label share link.
//
// DELETE /api/v1/labels/:uid/links/:link
func DeleteLabelLink(router *gin.RouterGroup) {
	router.DELETE("/labels/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		DeleteLink(c)
	})
}

// GetLabelLinks returns all share links for the given UID as JSON.
//
// GET /api/v1/labels/:uid/links
func GetLabelLinks(router *gin.RouterGroup) {
	router.GET("/labels/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		m, err := query.LabelByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}
*/
