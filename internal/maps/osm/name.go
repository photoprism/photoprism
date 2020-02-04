package osm

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

var labelTitles = map[string]string{
	"airport":        "Airport",
	"visitor center": "Visitor Center",
}

func (l Location) Name() (result string) {
	result = l.Category()

	if title, ok := labelTitles[result]; ok {
		title = strings.Replace(title, "%name%", l.LocName, 1)
		return title
	}

	if l.LocName != "" {
		result = l.LocName
	}

	result = strings.Replace(result, "_", " ", -1)

	if i := strings.Index(result, " - "); i > 1 {
		result = result[:i]
	}

	if i := strings.Index(result, ","); i > 1 {
		result = result[:i]
	}

	result = strings.SplitN(result, "/", 2)[0]

	return txt.Title(strings.TrimSpace(result))
}
