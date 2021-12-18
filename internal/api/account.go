package api

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/workers"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Namespaces for caching and logs.
const (
	accountFolder = "account-folder"
)

// GetAccount returns an account as JSON.
//
// GET /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func GetAccount(router *gin.RouterGroup) {
	router.GET("/accounts/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdUint(c.Param("id"))

		if m, err := query.AccountByID(id); err == nil {
			c.JSON(http.StatusOK, m)
		} else {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
		}
	})
}

// GetAccountFolders returns folders that belong to an account as JSON.
//
// GET /api/v1/accounts/:id/folders
//
// Parameters:
//   id: string Account ID as returned by the API
func GetAccountFolders(router *gin.RouterGroup) {
	router.GET("/accounts/:id/folders", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		start := time.Now()
		id := sanitize.IdUint(c.Param("id"))
		cache := service.FolderCache()
		cacheKey := fmt.Sprintf("%s:%d", accountFolder, id)

		if cacheData, ok := cache.Get(cacheKey); ok {
			cached := cacheData.(fs.FileInfos)

			log.Tracef("api: cache hit for %s [%s]", cacheKey, time.Since(start))

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
			log.Errorf("%s: %s", accountFolder, err.Error())
			Abort(c, http.StatusBadRequest, i18n.ErrConnectionFailed)
			return
		}

		cache.SetDefault(cacheKey, list)
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		c.JSON(http.StatusOK, list)
	})
}

// GET /api/v1/accounts/:id/share
//
// Parameters:
//   id: string Account ID as returned by the API
func ShareWithAccount(router *gin.RouterGroup) {
	router.POST("/accounts/:id/share", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionUpload)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		var f form.AccountShare

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		dst := f.Destination
		files, err := query.FilesByUID(f.Photos, 1000, 0)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var aliases = make(map[string]int)

		for _, file := range files {
			alias := path.Join(dst, file.ShareBase(0))
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.ShareBase(seq)
			}

			aliases[key] += 1

			entity.FirstOrCreateFileShare(entity.NewFileShare(file.ID, m.ID, alias))
		}

		workers.StartShare(service.Config())

		c.JSON(http.StatusOK, files)
	})
}

// POST /api/v1/accounts
func CreateAccount(router *gin.RouterGroup) {
	router.POST("/accounts", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionCreate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		var f form.Account

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if err := f.ServiceDiscovery(); err != nil {
			log.Error(err)
			Abort(c, http.StatusBadRequest, i18n.ErrConnectionFailed)
			return
		}

		m, err := entity.CreateAccount(f)

		if err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		event.SuccessMsg(i18n.MsgAccountCreated)

		c.JSON(http.StatusOK, m)
	})
}

// PUT /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func UpdateAccount(router *gin.RouterGroup) {
	router.PUT("/accounts/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		// 1) Init form with model values
		f, err := form.NewAccount(m)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// 2) Update form with values from request
		if err := c.BindJSON(&f); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		// 3) Save model with values from form
		if err := m.SaveForm(f); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		event.SuccessMsg(i18n.MsgAccountSaved)

		m, err = query.AccountByID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if m.AccSync {
			workers.StartSync(service.Config())
		}

		c.JSON(http.StatusOK, m)
	})
}

// DELETE /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func DeleteAccount(router *gin.RouterGroup) {
	router.DELETE("/accounts/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionDelete)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		id := sanitize.IdUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		if err := m.Delete(); err != nil {
			Error(c, http.StatusInternalServerError, err, i18n.ErrDeleteFailed)
			return
		}

		event.SuccessMsg(i18n.MsgAccountDeleted)

		c.JSON(http.StatusOK, m)
	})
}
