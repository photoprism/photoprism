package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GetConfigOptions returns backend config options.
//
//	@Summary	returns backend config options
//	@Id			GetConfigOptions
//	@Tags		Config, Settings
//	@Produce	json
//	@Success	200			{object}	config.Options
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/config/options [get]
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
//	@Summary	updates backend config options
//	@Id			SaveConfigOptions
//	@Tags		Config, Settings
//	@Produce	json
//	@Success	200					{object}	config.Options
//	@Failure	400,401,403,429,500	{object}	i18n.Response
//	@Param		options					body		config.Options	true	"properties to be updated (only submit values that should be changed)"
//	@Router		/api/v1/config/options [post]
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

			if err = yaml.Unmarshal(yamlData, v); err != nil {
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
		if err = fs.MkdirAll(filepath.Dir(fileName)); err != nil {
			log.Errorf("config: failed to create config path %s (%s)", filepath.Dir(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Write YAML data to file.
		if err = fs.WriteFile(fileName, yamlData); err != nil {
			log.Errorf("config: failed writing values to %s (%s)", clean.Log(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Reload options.
		if err = conf.Options().Load(fileName); err != nil {
			log.Warnf("config: failed loading values from %s (%s)", clean.Log(fileName), err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Set restart flag.
		mutex.Restart.Store(true)

		// Update package defaults.
		conf.Propagate()

		// Flush session cache and update client config.
		entity.FlushSessionCache()
		UpdateClientConfig()

		// Return updated config options.
		c.JSON(http.StatusOK, conf.Options())
	})
}
