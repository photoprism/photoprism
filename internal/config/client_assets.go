package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

type ClientAssets struct {
	BuildPath                 string `json:"-"`
	BaseUri                   string `json:"-"`
	AppCss                    string `json:"app.css"`
	AppJs                     string `json:"app.js"`
	ShareCss                  string `json:"share.css"`
	ShareJs                   string `json:"share.js"`
	SplashCss                 string `json:"splash.css"`
	SplashJs                  string `json:"splash.js"`
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
func NewClientAssets(buildPath, baseUri string) *ClientAssets {
	return &ClientAssets{BuildPath: buildPath, BaseUri: baseUri}
}

// Load loads the frontend assets from a webpack manifest file.
func (a *ClientAssets) Load(fileName string) error {
	jsonFile, err := os.ReadFile(filepath.Join(a.BuildPath, fileName))

	if err != nil {
		return err
	}

	return json.Unmarshal(jsonFile, a)
}

// AppCssUri returns the web app CSS URI.
func (a *ClientAssets) AppCssUri() string {
	if a.AppCss == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.AppCss)
}

// AppJsUri returns the web app JS URI.
func (a *ClientAssets) AppJsUri() string {
	if a.AppJs == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.AppJs)
}

// ShareCssUri returns the web sharing CSS URI.
func (a *ClientAssets) ShareCssUri() string {
	if a.ShareCss == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.ShareCss)
}

// ShareJsUri returns the web sharing JS URI.
func (a *ClientAssets) ShareJsUri() string {
	if a.ShareJs == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.ShareJs)
}

// SplashCssUri returns the splash screen CSS URI.
func (a *ClientAssets) SplashCssUri() string {
	if a.SplashCss == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.SplashCss)
}

// SplashCssFile returns the splash screen CSS filename.
func (a *ClientAssets) SplashCssFile() string {
	if a.SplashCss == "" {
		return ""
	}
	return a.SplashCss
}

// SplashCssFileContents returns the splash screen CSS file contents for embedding in HTML.
func (a *ClientAssets) SplashCssFileContents() template.CSS {
	return template.CSS(a.readFile(a.SplashCssFile()))
}

// SplashJsUri returns the splash screen JS URI.
func (a *ClientAssets) SplashJsUri() string {
	if a.ShareJs == "" {
		return ""
	}
	return fmt.Sprintf("%s/build/%s", a.BaseUri, a.SplashJs)
}

// SplashJsFile returns the splash screen JS filename.
func (a *ClientAssets) SplashJsFile() string {
	if a.ShareJs == "" {
		return ""
	}
	return a.ShareJs
}

// SplashJsFileContents returns the splash screen JS file contents for embedding in HTML.
func (a *ClientAssets) SplashJsFileContents() template.JS {
	if a.SplashJs == "" {
		return ""
	}
	return template.JS(a.readFile(a.SplashJs))
}

// readFile reads the file contents and returns them as string.
func (a *ClientAssets) readFile(fileName string) string {
	if fileName == "" {
		return ""
	} else if css, err := os.ReadFile(filepath.Join(a.BuildPath, fileName)); err != nil {
		return ""
	} else {
		return string(bytes.TrimSpace(css))
	}
}

// ClientAssets returns the frontend build assets.
func (c *Config) ClientAssets() *ClientAssets {
	result := NewClientAssets(c.BuildPath(), c.StaticUri())

	if err := result.Load("assets.json"); err != nil {
		log.Debugf("frontend: %s", err)
		log.Errorf("frontend: cannot read assets.json")
	}

	return result
}

// ClientManifestUri returns the frontend manifest.json URI.
func (c *Config) ClientManifestUri() string {
	return fmt.Sprintf("%s?%x", c.BaseUri("/manifest.json"), c.VersionChecksum())
}
