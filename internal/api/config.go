package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GetConfig returns client config values.
//
// GET /api/v1/config
func GetConfig(router *gin.RouterGroup) {
	router.GET("/config", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceConfig, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if s.User.IsGuest() {
			c.JSON(http.StatusOK, conf.GuestConfig())
		} else if s.User.IsRegistered() {
			c.JSON(http.StatusOK, conf.UserConfig())
		} else {
			c.JSON(http.StatusOK, conf.PublicConfig())
		}
	})
}

// GetConfigOptions returns backend config options.
//
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

// SaveConfigOptions updates backend config options.
//
// POST /api/v1/config/options
func SaveConfigOptions(router *gin.RouterGroup) {
	router.POST("/config/options", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceConfigOptions, acl.ActionUpdate)
		conf := service.Config()

		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		fileName := conf.OptionsYaml()

		if fileName == "" {
			log.Errorf("config: empty options.yml file path")
			AbortSaveFailed(c)
			return
		}

		type valueMap map[string]interface{}

		v := make(valueMap)

		if fs.FileExists(fileName) {
			yamlData, err := os.ReadFile(fileName)

			if err != nil {
				log.Errorf("config: failed loading values from %s (%s)", clean.Log(fileName), err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}

			if err := yaml.Unmarshal(yamlData, v); err != nil {
				log.Warnf("config: failed parsing values in %s (%s)", clean.Log(fileName), err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}
		}

		if err := c.BindJSON(&v); err != nil {
			log.Errorf("config: %s (bind json)", err)
			AbortBadRequest(c)
			return
		}

		yamlData, err := yaml.Marshal(v)

		if err != nil {
			log.Errorf("config: %s (marshal yaml)", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Make sure directory exists.
		if err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm); err != nil {
			log.Errorf("config: failed creating config path %s (%s)", filepath.Dir(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Write YAML data to file.
		if err := os.WriteFile(fileName, yamlData, os.ModePerm); err != nil {
			log.Errorf("config: failed writing values to %s (%s)", clean.Log(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Reload options.
		if err := conf.Options().Load(fileName); err != nil {
			log.Warnf("config: failed loading values from %s (%s)", clean.Log(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		conf.Propagate()

		UpdateClientConfig()

		event.InfoMsg(i18n.MsgSettingsSaved)

		c.JSON(http.StatusOK, conf.Options())
	})
}
