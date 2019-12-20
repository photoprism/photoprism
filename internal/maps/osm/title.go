package osm

import (
	"strings"

	"github.com/photoprism/photoprism/internal/util"
)

var labelTitles = map[string]string{
	"airport": "Airport",
	"visitor center": "Visitor Center",
}

func (o Location) Title() (result string) {
	result = o.Category()

	if title, ok := labelTitles[result]; ok {
		title = strings.Replace(title, "%name%", o.LocTitle, 1)
		return title
	}

	if o.LocTitle != "" {
		result = o.LocTitle
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
