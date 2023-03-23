package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Namespaces for caching and logs.
const (
	serviceFolder = "service-folder"
)

// GetService returns an account as JSON.
//
// GET /api/v1/services/:id
func GetService(router *gin.RouterGroup) {
	router.GET("/services/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionView)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		id := clean.IdUint(c.Param("id"))

		if m, err := query.AccountByID(id); err == nil {
			c.JSON(http.StatusOK, m)
		} else {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
		}
	})
}

// GetServiceFolders returns folders that belong to an account as JSON.
//
// GET /api/v1/services/:id/folders
func GetServiceFolders(router *gin.RouterGroup) {
	router.GET("/services/:id/folders", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionView)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		start := time.Now()
		id := clean.IdUint(c.Param("id"))
		cache := get.FolderCache()
		cacheKey := fmt.Sprintf("%s:%d", serviceFolder, id)

		if cacheData, ok := cache.Get(cacheKey); ok {
			cached := cacheData.(fs.FileInfos)

			log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

			c.JSON(http.StatusOK, cached)
			return
		}

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		list, err := m.Directories()

		if err != nil {
			log.Errorf("%s: %s", serviceFolder, err.Error())
			Abort(c, http.StatusBadRequest, i18n.ErrConnectionFailed)
			return
		}

		cache.SetDefault(cacheKey, list)
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		c.JSON(http.StatusOK, list)
	})
}

// AddService creates a new remote account configuration.
//
// POST /api/v1/services
func AddService(router *gin.RouterGroup) {
	router.POST("/services", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionCreate)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		var f form.Service

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if err := f.Discovery(); err != nil {
			log.Error(err)
			Abort(c, http.StatusBadRequest, i18n.ErrConnectionFailed)
			return
		}

		m, err := entity.AddService(f)

		if err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}

// UpdateService updates a remote account configuration.
//
// PUT /api/v1/services/:id
func UpdateService(router *gin.RouterGroup) {
	router.PUT("/services/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		id := clean.IdUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		// 1) Init form with model values
		f, err := form.NewService(m)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// 2) Update form with values from request
		if err = c.BindJSON(&f); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		// 3) Save model with values from form
		if err = m.SaveForm(f); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		m, err = query.AccountByID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if m.AccSync {
			workers.RunSync(get.Config())
		}

		c.JSON(http.StatusOK, m)
	})
}

// DeleteService removes a remote account configuration.
//
// DELETE /api/v1/services/:id
func DeleteService(router *gin.RouterGroup) {
	router.DELETE("/services/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		id := clean.IdUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		if err := m.Delete(); err != nil {
			Error(c, http.StatusInternalServerError, err, i18n.ErrDeleteFailed)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}
