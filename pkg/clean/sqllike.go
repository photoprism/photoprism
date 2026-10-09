package clean

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// SqlLikeEscape is the escape character of the LIKE conditions built with SqlLikeCond.
//
// Not a backslash: MySQL reads one inside a string literal as an escape while SQLite does not, so
// the ESCAPE clause itself cannot be written the same way for both. Nothing needs escaping in "!".
const SqlLikeEscape = "!"

// sqlLikeEscaper escapes the LIKE wildcards and the escape character itself.
var sqlLikeEscaper = strings.NewReplacer(SqlLikeEscape, SqlLikeEscape+SqlLikeEscape, "%", SqlLikeEscape+"%", "_", SqlLikeEscape+"_")

// SqlLike escapes the LIKE wildcards in s, so that it matches literally in a SqlLikeCond condition.
func SqlLike(s string) string {
	return sqlLikeEscaper.Replace(s)
}

// sqlNoMatch returns a condition that is never true and has n placeholders, so the arguments a caller
// passes still bind when a column is rejected. PostgreSQL needs a type for each placeholder.
func sqlNoMatch(dialect string, n int) string {
	if dialect == dsn.DialectPostgreSQL {
		return "(1 = 0" + strings.Repeat(" AND CAST(? AS VARCHAR) IS NULL", n) + ")"
	}

	return "(1 = 0" + strings.Repeat(" AND ? IS NULL", n) + ")"
}

// sqlText returns col as text: binary columns are decoded on PostgreSQL, so that the text functions
// and patterns of the other conditions apply to them.
func sqlText(dialect string, binary bool, col string) string {
	if binary && dialect == dsn.DialectPostgreSQL {
		return fmt.Sprintf("convert_from(%s, 'UTF8')", col)
	}

	return col
}

// SqlLikeCond returns "col LIKE ? ESCAPE '!'" for the SQL dialect, or a condition that matches nothing
// if col is not a plain column name. Set binary for VARBINARY columns. PostgreSQL compares text columns
// with ILIKE and decoded binary columns with LIKE, so letter case matters as it does on MariaDB.
func SqlLikeCond(dialect string, binary bool, col string) string {
	if SqlColumn(col) == "" {
		return sqlNoMatch(dialect, 1)
	}

	op := "LIKE"

	if dialect == dsn.DialectPostgreSQL && !binary {
		op = "ILIKE"
	}

	return fmt.Sprintf("%s %s ? ESCAPE '%s'", sqlText(dialect, binary, col), op, SqlLikeEscape)
}

// SqlLikeAny returns a condition that matches if any of the columns is LIKE the bound value, with one
// placeholder per column, or a condition that matches nothing if a column is not a plain column name.
// All columns must be binary or all must be text; combine two conditions otherwise.
func SqlLikeAny(dialect string, binary bool, cols ...string) string {
	conds := make([]string, len(cols))

	for i, col := range cols {
		if SqlColumn(col) == "" {
			return sqlNoMatch(dialect, len(cols))
		}

		conds[i] = SqlLikeCond(dialect, binary, col)
	}

	if len(conds) == 0 {
		return sqlNoMatch(dialect, 0)
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}

// SqlLikeExpr returns an SQL expression that escapes the LIKE wildcards in the value of col, like
// SqlLike does for a bound value, or NULL, which matches nothing, if col is not a plain column name.
// The escape character is replaced innermost, so the ones the outer calls add are not escaped again.
func SqlLikeExpr(dialect string, binary bool, col string) string {
	if SqlColumn(col) == "" {
		return "NULL"
	}

	e := SqlLikeEscape

	return fmt.Sprintf("REPLACE(REPLACE(REPLACE(%s, '%s', '%s'), '%%', '%s%%'), '_', '%s_')", sqlText(dialect, binary, col), e, e+e, e, e)
}

// SqlPrefixCond returns a condition matching the values of col that start with a prefix, compared
// case-sensitively on every driver, or a condition that matches nothing if col is not a plain column
// name. Bind the arguments SqlPrefixArgs returns; the LIKE clause keeps an index usable for the lookup.
func SqlPrefixCond(dialect string, binary bool, col string) string {
	if SqlColumn(col) == "" {
		return sqlNoMatch(dialect, 3)
	}

	text := sqlText(dialect, binary, col)

	return fmt.Sprintf("(%s LIKE ? ESCAPE '%s' AND SUBSTR(%s, 1, LENGTH(?)) = ?)", text, SqlLikeEscape, text)
}

// SqlPrefixArgs returns the arguments of a SqlPrefixCond condition for the given prefix.
func SqlPrefixArgs(prefix string) []any {
	return []any{SqlLike(prefix) + "%", prefix, prefix}
}
