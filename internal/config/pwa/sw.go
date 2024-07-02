package pwa

// Serviceworker represents serviceworker options for the PWA manifest.
type Serviceworker struct {
	Src      string `json:"src"`
	Scope    string `json:"scope"`
	UseCache bool   `json:"use_cache"`
}
