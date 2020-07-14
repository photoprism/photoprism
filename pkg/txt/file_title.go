package txt

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

var FileTitleRegexp = regexp.MustCompile("[\\p{L}\\-,':]{2,}")

// FileTitle returns the string with the first characters of each word converted to uppercase.
func FileTitle(s string) string {
	s = fs.BasePrefix(s, true)

	if len(s) < 3 {
		return ""
	}

	words := FileTitleRegexp.FindAllString(s, -1)
	var result []string

	found := 0

	for _, w := range words {
		w = strings.ToLower(w)

		if len(w) < 3 && found == 0 {
			continue
		}

		if _, ok := StopWords[w]; ok && found == 0 {
			continue
		}

		if UnknownWord(w) {
			continue
		}

		result = append(result, w)

		found++

		if found > 10 {
			break
		}
	}

	if found == 0 {
		return ""
	}

	title := strings.Join(result, " ")

	title = strings.ReplaceAll(title, "--", " / ")
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.ReplaceAll(title, "  ", " ")

	if len(title) <= 4 {
		return ""
	}

	return Title(title)
}
