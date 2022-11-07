package txt

import (
	"strings"
)

// Name represents the components of a full name.
type Name struct {
	Title  string
	Given  string
	Middle string
	Family string
	Suffix string
	Nick   string
}

// ParseName parses a full name and returns the components.
func ParseName(full string) Name {
	name := Name{}
	name.Parse(full)
	return name
}

// Parse tries to parse a full name.
func (name *Name) Parse(full string) {
	if full == "" {
		return
	}

	for _, w := range KeywordsRegexp.FindAllString(full, -1) {
		w = strings.Trim(w, "- '")

		if w == "" || len(w) < 2 && IsLatin(w) {
			continue
		}

		l := strings.ToLower(w)

		if IsNameTitle[l] {
			name.Title = AppendName(name.Title, w)
		} else if IsNameSuffix[l] {
			name.Suffix = AppendName(name.Suffix, w)
		} else if name.Given == "" {
			name.Given = w
		} else {
			name.Family = AppendName(name.Family, w)
		}
	}

	return
}
