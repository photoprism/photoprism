package osm

import (
	"strings"

	"github.com/photoprism/photoprism/internal/util"
)

var labelTitles = map[string]string{
	"airport":        "Airport",
	"visitor center": "Visitor Center",
}

func (o Location) Name() (result string) {
	result = o.Category()

	if title, ok := labelTitles[result]; ok {
		title = strings.Replace(title, "%name%", o.LocName, 1)
		return title
	}

	if o.LocName != "" {
		result = o.LocName
	}

	result = strings.Replace(result, "_", " ", -1)

	if i := strings.Index(result, " - "); i > 1 {
		result = result[:i]
	}

	if i := strings.Index(result, ","); i > 1 {
		result = result[:i]
	}

	return util.Title(strings.TrimSpace(result))
}
