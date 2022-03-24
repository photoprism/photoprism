package search

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

// SqlLike escapes a string for use in an SQL query.
func SqlLike(s string) string {
	return strings.Trim(sanitize.SqlString(s), " |&*%")
}
