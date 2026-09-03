package sortby

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// Find PostgreSQL column names in order by strings that need collate statements.
var PostgreSQLOrderColumnsMatch = regexp.MustCompile(`(?:^|\.| )(?:acc_name|album_category|album_title|display_name|keyword|marker_name|photo_title|subj_name|user_name)(?: |,|$)`)

// This is the list of From and To replacements for PostgreSQL order by strings.
var PostgreSQLOrderColumnsSlice = []string{
	`acc_name`, `acc_name COLLATE "caseinsensitive"`,
	`album_category`, `album_category COLLATE "caseinsensitive"`,
	`album_title`, `album_title COLLATE "caseinsensitive"`,
	`display_name`, `display_name COLLATE "caseinsensitive"`,
	`keyword`, `keyword COLLATE "caseinsensitive"`,
	`marker_name`, `marker_name COLLATE "caseinsensitive"`,
	`photo_title`, `photo_title COLLATE "caseinsensitive"`,
	`subj_name`, `subj_name COLLATE "caseinsensitive"`,
	`user_name`, `user_name COLLATE "caseinsensitive"`}

// OrderReplacer replaces "ASC" with "DESC" and "DESC" with "ASC", "FIRST" with "LAST" and "LAST" with "FIRST"
var PostgreSQLOrderByReplacer = strings.NewReplacer(PostgreSQLOrderColumnsSlice[0], PostgreSQLOrderColumnsSlice[1],
	PostgreSQLOrderColumnsSlice[2], PostgreSQLOrderColumnsSlice[3],
	PostgreSQLOrderColumnsSlice[4], PostgreSQLOrderColumnsSlice[5],
	PostgreSQLOrderColumnsSlice[6], PostgreSQLOrderColumnsSlice[7],
	PostgreSQLOrderColumnsSlice[8], PostgreSQLOrderColumnsSlice[9],
	PostgreSQLOrderColumnsSlice[10], PostgreSQLOrderColumnsSlice[11],
	PostgreSQLOrderColumnsSlice[12], PostgreSQLOrderColumnsSlice[13],
	PostgreSQLOrderColumnsSlice[14], PostgreSQLOrderColumnsSlice[15],
	PostgreSQLOrderColumnsSlice[16], PostgreSQLOrderColumnsSlice[17])

// DialectOrderByFix updates order by strings to comply with requirements for specific database dialects.
func DialectOrderByFix(s string, dialect string) string {
	if dialect == dsn.DialectPostgreSQL {
		return PostgreSQLOrderColumnsMatch.ReplaceAllStringFunc(s, PostgreSQLOrderByFix)
	} else {
		return s
	}
}

// PostgreSQLOrderByFix appends collation to specific column names in the subset of an order by statement
func PostgreSQLOrderByFix(s string) string {
	return PostgreSQLOrderByReplacer.Replace(s)
}
