package pwa

// Config represents progressive web app manifest config values.
type Config struct {
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mode        string `json:"mode"`
	BaseUri     string `json:"baseUri"`
	StaticUri   string `json:"staticUri"`
}
