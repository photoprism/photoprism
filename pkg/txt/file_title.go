package txt

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

var FileTitleRegexp = regexp.MustCompile("[\\p{L}\\d,':&+!?！]{1,}[\\-]{0,}|( [&+] )?")

// FileTitle returns the string with the first characters of each word converted to uppercase.
func FileTitle(s string) string {
	s = fs.BasePrefix(s, true)

	if len(s) < 3 && IsASCII(s) {
		return ""
	}

	words := FileTitleRegexp.FindAllString(s, -1)
	var result []string

	found := 0

	for _, w := range words {
		// Ignore purely numeric values, such as those in timestamps.
		if found == 0 && IsDateNumber(w) {
			continue
		} else if IsNumeric(w) {
			continue
		}

		// Ignore ASCII strings < 3 characters at the beginning.
		if IsASCII(w) && (len(w) < 3 && found == 0 || len(w) == 1) {
			continue
		}

		// Ignore short strings that are not a known word.
		if UnknownWord(w) {
			continue
		}

		result = append(result, w)

		found++

		if found > 16 {
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

	// Remove small words from title ending.
	for w := range SmallWords {
		title = strings.TrimSuffix(title, " "+w)
	}

	if len(title) <= 4 && IsASCII(title) || IsDateNumber(title) {
		return ""
	}

	return Title(title)
}
