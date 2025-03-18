package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// Namespaces for caching and logs.
const (
	serviceFolder = "service-folder"
)

// GetService returns an account as JSON.
//
//	@Summary	returns the specified remote service account configuration as JSON
//	@Id			GetService
//	@Tags		Services
//	@Produce	json
//	@Success	200				{object}	entity.Service
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		id				path		string	true	"service id"
//	@Router		/api/v1/services/{id} [get]
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

// GetServiceFolders returns folders that belong to an account.
//
//	@Summary	returns folders that belong to a remote service account
//	@Id			GetServiceFolders
//	@Tags		Services
//	@Produce	json
//	@Success	200				{object}	[]object
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		id				path		string	true	"service id"
//	@Router		/api/v1/services/{id}/folders [get]
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
//	@Summary	creates a new remote service account configuration
//	@Id			AddService
//	@Tags		Services
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Service
//	@Failure	400,401,403,429	{object}	i18n.Response
//	@Param		service				body		form.Service	true	"properties of the service to be created"
//	@Router		/api/v1/services [post]
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

		var frm form.Service

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c)
			return
		}

		if err := frm.Discovery(); err != nil {
			log.Error(err)
			Abort(c, http.StatusBadRequest, i18n.ErrConnectionFailed)
			return
		}

		m, err := entity.AddService(frm)

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
//	@Summary	updates a remote account configuration
//	@Id			UpdateService
//	@Tags		Services
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Service
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		id						path		string		true	"service id"
//	@Param		service					body		form.Service	true	"properties to be updated (only submit values that should be changed)"
//	@Router		/api/v1/services/{id} [put]
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
		frm, err := form.NewService(m)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// 2) Update form with values from request
		if err = c.BindJSON(&frm); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		// 3) Save model with values from form
		if err = m.SaveForm(frm); err != nil {
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
//	@Summary	removes a remote service account configuration
//	@Id			DeleteService
//	@Tags		Services
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Service
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		id					path		string	true	"service id"
//	@Router		/api/v1/services/{id} [delete]
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
