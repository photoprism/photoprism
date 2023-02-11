package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ClientAssets struct {
	BaseUri                   string `json:"-"`
	AppCss                    string `json:"app.css"`
	AppJs                     string `json:"app.js"`
	ShareCss                  string `json:"share.css"`
	ShareJs                   string `json:"share.js"`
	MaterialIconsRegularTtf   string `json:"MaterialIcons-Regular.ttf"`
	MaterialIconsRegularWoff  string `json:"MaterialIcons-Regular.woff"`
	MaterialIconsRegularEot   string `json:"MaterialIcons-Regular.eot"`
	MaterialIconsRegularWoff2 string `json:"MaterialIcons-Regular.woff2"`
	OfflineServiceworker      string `json:"__offline_serviceworker"`
	DefaultSkinSvg            string `json:"default-skin.svg"`
	PreloaderGif              string `json:"preloader.gif"`
	DefaultSkinPng            string `json:"default-skin.png"`
}

// NewClientAssets creates a new ClientAssets instance.
func NewClientAssets(baseUri string) ClientAssets {
	return ClientAssets{BaseUri: baseUri}
}

// Load loads the frontend assets from a webpack manifest file.
func (a *ClientAssets) Load(fileName string) error {
	jsonFile, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	return json.Unmarshal(jsonFile, a)
}

// AppCssUri returns the web app stylesheet URI.
func (a *ClientAssets) AppCssUri() string {
	if a.AppCss == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.AppCss)
}

// AppJsUri returns the web app javascript URI.
func (a *ClientAssets) AppJsUri() string {
	if a.AppJs == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.AppJs)
}

// ShareCssUri returns the web sharing stylesheet URI.
func (a *ClientAssets) ShareCssUri() string {
	if a.ShareCss == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.ShareCss)
}

// ShareJsUri returns the web sharing javascript URI.
func (a *ClientAssets) ShareJsUri() string {
	if a.ShareJs == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.ShareJs)
}

// ClientAssets returns the frontend build assets.
func (c *Config) ClientAssets() ClientAssets {
	result := NewClientAssets(c.StaticUri())

	if err := result.Load(filepath.Join(c.BuildPath(), "assets.json")); err != nil {
		log.Debugf("frontend: %s", err)
		log.Errorf("frontend: cannot read assets.json")
	}

	return result
}

// ClientManifestUri returns the frontend manifest.json URI.
func (c *Config) ClientManifestUri() string {
	return fmt.Sprintf("%s?%x", c.BaseUri("/manifest.json"), c.VersionChecksum())
}
