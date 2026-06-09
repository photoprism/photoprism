package pwa

import "strings"

// Url represents a URL with a name.
type Url struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// Urls represents a set of URLs.
type Urls []Url

// scopeRelativeUrl returns a URL relative to the app scope when possible.
func scopeRelativeUrl(scope string, uri string) string {
	if uri = strings.TrimSpace(uri); uri == "" {
		return ""
	}

	if strings.HasPrefix(uri, "//") || strings.Contains(uri, "://") {
		return uri
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	if scope = strings.TrimSpace(scope); scope == "" {
		scope = "/"
	}

	if !strings.HasPrefix(scope, "/") {
		scope = "/" + scope
	}

	if !strings.HasSuffix(scope, "/") {
		scope += "/"
	}

	if scope == "/" {
		return strings.TrimPrefix(uri, "/")
	}

	if strings.HasPrefix(uri, scope) {
		return strings.TrimPrefix(uri, scope)
	}

	return uri
}

// StartUrl returns the manifest start_url relative to the app scope when possible.
func StartUrl(scope string, frontendUri string) string {
	startUrl := scopeRelativeUrl(scope, frontendUri)

	if startUrl == "" {
		return ""
	}

	if strings.HasPrefix(startUrl, "/") || strings.HasPrefix(startUrl, "//") || strings.Contains(startUrl, "://") {
		return startUrl
	}

	if strings.HasPrefix(startUrl, "./") || strings.HasPrefix(startUrl, "../") {
		return startUrl
	}

	return "./" + startUrl
}

// shortcutUrl joins a route with the frontend base URI using scope-relative URLs when possible.
func shortcutUrl(scope string, frontendUri string, route string) string {
	base := scopeRelativeUrl(scope, frontendUri)
	route = strings.TrimLeft(route, "/")

	if base == "" {
		return route
	}

	if strings.HasPrefix(base, "/") || strings.HasPrefix(base, "//") || strings.Contains(base, "://") {
		return strings.TrimRight(base, "/") + "/" + route
	}

	return strings.TrimRight(base, "/") + "/" + route
}

// Shortcuts specifies links to key tasks or pages within the web application,
// see https://developer.mozilla.org/en-US/docs/Web/Manifest/Reference/shortcuts.
func Shortcuts(scope string, frontendUri string) Urls {
	return Urls{
		{
			Name: "Search",
			Url:  shortcutUrl(scope, frontendUri, "browse"),
		},
		{
			Name: "Albums",
			Url:  shortcutUrl(scope, frontendUri, "albums"),
		},
		{
			Name: "Places",
			Url:  shortcutUrl(scope, frontendUri, "places"),
		},
		{
			Name: "Settings",
			Url:  shortcutUrl(scope, frontendUri, "settings"),
		},
	}
}

// PhotoPrism specifies the developer contact URL.
var PhotoPrism = Url{
	"PhotoPrism",
	"https://www.photoprism.app/",
}
