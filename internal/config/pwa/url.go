package pwa

// Url represents a URL with a name.
type Url struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// Urls represents a set of URLs.
type Urls []Url

// Shortcuts specifies links to key tasks or pages within the web application,
// see https://developer.mozilla.org/en-US/docs/Web/Manifest/Reference/shortcuts.
func Shortcuts(baseUri string) Urls {
	return Urls{
		{
			Name: "Search",
			Url:  baseUri + "library/browse",
		},
		{
			Name: "Albums",
			Url:  baseUri + "library/albums",
		},
		{
			Name: "Places",
			Url:  baseUri + "library/places",
		},
		{
			Name: "Settings",
			Url:  baseUri + "library/settings",
		},
	}
}

// PhotoPrism specifies the developer contact URL.
var PhotoPrism = Url{
	"PhotoPrism",
	"https://www.photoprism.app/",
}
