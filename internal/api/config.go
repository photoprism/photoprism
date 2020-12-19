package api

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
	"gopkg.in/yaml.v2"
)

// GET /api/v1/config
func GetConfig(router *gin.RouterGroup) {
	router.GET("/config", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceConfig, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if s.User.Guest() {
			c.JSON(http.StatusOK, conf.GuestConfig())
		} else if s.User.Registered() {
			c.JSON(http.StatusOK, conf.UserConfig())
		} else {
			c.JSON(http.StatusOK, conf.PublicConfig())
		}
	})
}

// GET /api/v1/config/options
func GetConfigOptions(router *gin.RouterGroup) {
	router.GET("/config/options", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceConfigOptions, acl.ActionRead)
		conf := service.Config()

		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		c.JSON(http.StatusOK, conf.Options())
	})
}

// POST /api/v1/config/options
func SaveConfigOptions(router *gin.RouterGroup) {
	router.POST("/config/options", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceConfigOptions, acl.ActionUpdate)
		conf := service.Config()

		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		fileName := conf.ConfigFile()

		if fileName == "" {
			log.Errorf("options: empty config file name")
			AbortSaveFailed(c)
			return
		}

		type valueMap map[string]interface{}

		v := make(valueMap)

		if fs.FileExists(fileName) {
			yamlData, err := ioutil.ReadFile(fileName)

			if err != nil {
				log.Errorf("options: %s", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}

			if err := yaml.Unmarshal(yamlData, v); err != nil {
				log.Errorf("options: %s", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}
		}

		if err := c.BindJSON(&v); err != nil {
			log.Errorf("options: %s", err)
			AbortBadRequest(c)
			return
		}

		yamlData, err := yaml.Marshal(v)

		if err != nil {
			log.Errorf("options: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Make sure directory exists.
		if err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm); err != nil {
			log.Errorf("options: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Write YAML data to file.
		if err := ioutil.WriteFile(fileName, yamlData, os.ModePerm); err != nil {
			log.Errorf("options: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		if err := conf.Options().Load(fileName); err != nil {
			log.Errorf("options: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		conf.Propagate()

		UpdateClientConfig()

		log.Infof(i18n.Msg(i18n.MsgSettingsSaved))

		c.JSON(http.StatusOK, conf.Options())
	})
}
