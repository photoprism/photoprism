package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GET /api/v1/accounts
func GetAccounts(router *gin.RouterGroup) {
	router.GET("/accounts", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAccounts, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.AccountSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := query.AccountSearch(f)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// TODO c.Header("X-Count", strconv.Itoa(count))
		c.Header("X-Limit", strconv.Itoa(f.Count))
		c.Header("X-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

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

		id := ParseUint(c.Param("id"))

		if m, err := query.AccountByID(id); err == nil {
			c.JSON(http.StatusOK, m)
		} else {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
		}
	})
}

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

		start := time.Now()
		id := ParseUint(c.Param("id"))
		cache := service.Cache()
		cacheKey := fmt.Sprintf("account-folders:%d", id)

		if cacheData, err := cache.Get(cacheKey); err == nil {
			var cached fs.FileInfos

			if err := json.Unmarshal(cacheData, &cached); err != nil {
				log.Errorf("account-folders: %s", err)
			} else {
				log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))
				c.JSON(http.StatusOK, cached)
				return
			}
		}

		m, err := query.AccountByID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrAccountNotFound)
			return
		}

		list, err := m.Directories()

		if err != nil {
			log.Errorf("account-folders: %s", err.Error())
			Abort(c, http.StatusBadRequest, i18n.ErrConnectionFailed)
			return
		}

		if c, err := json.Marshal(list); err == nil {
			logError("account-folders", cache.Set(cacheKey, c))
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

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

		id := ParseUint(c.Param("id"))

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

		for _, file := range files {
			dstFileName := dst + "/" + file.ShareFileName()

			fileShare := entity.NewFileShare(file.ID, m.ID, dstFileName)
			entity.FirstOrCreateFileShare(fileShare)
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

		log.Debugf("account: creating %+v %+v", f, m)

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

		id := ParseUint(c.Param("id"))

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

		id := ParseUint(c.Param("id"))

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
