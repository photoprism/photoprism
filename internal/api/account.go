package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/accounts
func GetAccounts(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/accounts", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.AccountSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := query.AccountSearch(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
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
func GetAccount(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/accounts/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))

		if m, err := query.AccountByID(id); err == nil {
			c.JSON(http.StatusOK, m)
		} else {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
		}
	})
}

// GET /api/v1/accounts/:id/dirs
//
// Parameters:
//   id: string Account ID as returned by the API
func GetAccountDirs(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/accounts/:id/dirs", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
			return
		}

		list, err := m.Directories()

		if err != nil {
			log.Errorf("account: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusNotFound, ErrConnectionFailed)
			return
		}

		c.JSON(http.StatusOK, list)
	})
}

// GET /api/v1/accounts/:id/share
//
// Parameters:
//   id: string Account ID as returned by the API
func ShareWithAccount(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/accounts/:id/share", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
			return
		}

		var f form.AccountShare

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		dst := f.Destination
		files, err := query.FilesByUID(f.Photos, 1000, 0)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		for _, file := range files {
			dstFileName := dst + "/" + file.ShareFileName()

			fileShare := entity.NewFileShare(file.ID, m.ID, dstFileName)
			entity.FirstOrCreateFileShare(fileShare)
		}

		workers.StartShare(conf)

		c.JSON(http.StatusOK, files)
	})
}

// POST /api/v1/accounts
func CreateAccount(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/accounts", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.Account

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if err := f.ServiceDiscovery(); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		m, err := entity.CreateAccount(f)

		log.Debugf("create account: %+v %+v", f, m)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		event.Success("account created")

		c.JSON(http.StatusOK, m)
	})
}

// PUT /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func UpdateAccount(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/accounts/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		// 1) Init form with model values
		f, err := form.NewAccount(m)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		// 2) Update form with values from request
		if err := c.BindJSON(&f); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrFormInvalid)
			return
		}

		// 3) Save model with values from form
		if err := m.SaveForm(f); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		event.Success("account saved")

		m, err = query.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}

// DELETE /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func DeleteAccount(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/accounts/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))

		m, err := query.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
			return
		}

		if err := m.Delete(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, txt.UcFirst(err.Error()))
			return
		}

		event.Success("account deleted")

		c.JSON(http.StatusOK, m)
	})
}
