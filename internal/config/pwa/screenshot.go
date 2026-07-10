package pwa

import "fmt"

// Screenshots represents a list of app store style install-prompt screenshots.
type Screenshots []Screenshot

// Screenshot represents a single install-prompt screenshot.
type Screenshot struct {
	Src        string `json:"src"`
	Sizes      string `json:"sizes,omitempty"`
	Type       string `json:"type,omitempty"`
	FormFactor string `json:"form_factor,omitempty"`
}

// NewScreenshots creates the install-prompt screenshots based on the config provided.
// Browsers show a richer, app-store-style install dialog when wide and narrow form factors are present.
func NewScreenshots(c Config) Screenshots {
	staticUri := c.StaticUri

	return Screenshots{
		{
			Src:        fmt.Sprintf("%s/img/screenshots/wide.jpg", staticUri),
			Sizes:      "1280x900",
			Type:       "image/jpeg",
			FormFactor: "wide",
		},
		{
			Src:        fmt.Sprintf("%s/img/screenshots/narrow.jpg", staticUri),
			Sizes:      "375x667",
			Type:       "image/jpeg",
			FormFactor: "narrow",
		},
	}
}
