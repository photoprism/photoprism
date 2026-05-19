package pwa

// Config represents progressive web app manifest config values.
type Config struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	DefaultLocale string `json:"defaultLocale"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Mode          string `json:"mode"`
	BaseUri       string `json:"baseUri"`
	FrontendUri   string `json:"frontendUri"`
	StaticUri     string `json:"staticUri"`
	SiteUrl       string `json:"siteUrl"`
	CdnUrl        string `json:"cdnUrl"`
	ThemeUri      string `json:"themeUri"`
	ThemePath     string `json:"themePath"`
}
