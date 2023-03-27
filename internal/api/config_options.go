package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GetConfigOptions returns backend config options.
//
// GET /api/v1/config/options
func GetConfigOptions(router *gin.RouterGroup) {
	router.GET("/config/options", func(c *gin.Context) {
		s := Auth(c, acl.ResourceConfig, acl.AccessAll)
		conf := get.Config()

		// Abort if permission was not granted.
		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
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
		s := Auth(c, acl.ResourceConfig, acl.ActionManage)
		conf := get.Config()

		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
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
		if err := os.MkdirAll(filepath.Dir(fileName), fs.ModeDir); err != nil {
			log.Errorf("config: failed creating config path %s (%s)", filepath.Dir(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Write YAML data to file.
		if err := os.WriteFile(fileName, yamlData, fs.ModeFile); err != nil {
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

		// Set restart flag.
		mutex.Restart.Store(true)

		// Propagate changes.
		conf.Propagate()

		// Flush session cache and update client config.
		entity.FlushSessionCache()
		UpdateClientConfig()

		// Return updated config options.
		c.JSON(http.StatusOK, conf.Options())
	})
}
