package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestGetConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetClientConfig(router)

		r := PerformRequest(app, "GET", "/api/v1/config")
		val := gjson.Get(r.Body.String(), "flags")

		if conf.Develop() {
			assert.Equal(t, "public debug test sponsor develop experimental settings", val.String())
		} else {
			assert.Equal(t, "public debug test sponsor experimental settings", val.String())
		}

		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestGetConfigOptions(t *testing.T) {
	t.Run("Forbidden", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetConfigOptions(router)

		r := PerformRequest(app, "GET", "/api/v1/config/options")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestSaveConfigOptions(t *testing.T) {
	t.Run("Forbidden", func(t *testing.T) {
		app, router, _ := NewApiTest()

		SaveConfigOptions(router)

		r := PerformRequest(app, "POST", "/api/v1/config/options")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()

		SaveConfigOptions(router)

		prepareConfigOptionsSuccessTest(t, conf)

		authToken := AuthenticateAdmin(app, router)

		tempCfg := t.TempDir()
		originalConfigPath := conf.Options().ConfigPath
		originalOptionsYaml := conf.Options().OptionsYaml

		t.Cleanup(func() {
			conf.Options().ConfigPath = originalConfigPath
			conf.Options().OptionsYaml = originalOptionsYaml
		})

		conf.Options().ConfigPath = tempCfg
		conf.Options().OptionsYaml = filepath.Join(tempCfg, "options.yml")

		seed := map[string]any{
			"Existing":        "value",
			"SiteUrl":         "https://old.example/",
			"HttpCachePublic": false,
		}

		b, err := yaml.Marshal(seed)

		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(conf.OptionsYaml(), b, fs.ModeFile))

		r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/config/options", `{"SiteUrl":"https://photos.example.com/","HttpCachePublic":true}`, authToken)

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "https://photos.example.com/", gjson.Get(r.Body.String(), "SiteUrl").String())
		assert.True(t, gjson.Get(r.Body.String(), "HttpCachePublic").Bool())

		optionsData, readErr := os.ReadFile(conf.OptionsYaml())
		assert.NoError(t, readErr)

		var merged map[string]any
		assert.NoError(t, yaml.Unmarshal(optionsData, &merged))
		assert.Equal(t, "value", merged["Existing"])
		assert.Equal(t, "https://photos.example.com/", merged["SiteUrl"])
		assert.Equal(t, true, merged["HttpCachePublic"])
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, conf := NewApiTest()

		SaveConfigOptions(router)

		prepareConfigOptionsSuccessTest(t, conf)

		authToken := AuthenticateAdmin(app, router)
		body := `{"SiteUrl":"https://photos.example.com/","LogLevel":"` + strings.Repeat("a", int(MaxSettingsRequestBytes)) + `"}`
		r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/config/options", body, authToken)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}

// prepareConfigOptionsSuccessTest normalizes shared config state so the success
// path is deterministic when API tests mutate singleton options.
func prepareConfigOptionsSuccessTest(t *testing.T, conf *config.Config) {
	t.Helper()

	originalOptions := *conf.Options()
	t.Cleanup(func() {
		*conf.Options() = originalOptions
	})

	conf.Options().AuthMode = config.AuthModePasswd
	conf.Options().Public = false
	conf.Options().Demo = false
	conf.Options().DisableSettings = false
}
