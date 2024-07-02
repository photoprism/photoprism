package migrate

import "strings"

type QueryErr map[string][]string

// Matches checks if there is a match for the specified query and error string.
func (m QueryErr) Matches(query, err string) bool {
	query = strings.ToLower(query)
	err = strings.ToLower(err)
	for substr, e := range m {
		if strings.Contains(query, substr) {
			for _, s := range e {
				if strings.Contains(err, s) {
					return true
				}
			}
		}
	}

	return false
}

var IgnoreErr = QueryErr{
	"rename":       {"no such", "already exists"},
	"replace":      {"no such", "exist", "exists"},
	" ignore ":     {"no such", "exist", "exists"},
	"drop index ":  {"drop"},
	"drop table ":  {"drop"},
	"alter table ": {"duplicate"},
}
