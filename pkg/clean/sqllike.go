package clean

import (
	"fmt"
	"strings"
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
// passes still bind when a column is rejected.
func sqlNoMatch(n int) string {
	return "(1 = 0" + strings.Repeat(" AND ? IS NULL", n) + ")"
}

// SqlLikeCond returns "col LIKE ? ESCAPE '!'", or a condition that matches nothing if col is not a
// plain column name. The drivers disagree on the default escape character, so a SqlLike value needs
// this ESCAPE clause.
func SqlLikeCond(col string) string {
	if SqlColumn(col) == "" {
		return sqlNoMatch(1)
	}

	return fmt.Sprintf("%s LIKE ? ESCAPE '%s'", col, SqlLikeEscape)
}

// SqlLikeAny returns a condition that matches if any of the columns is LIKE the bound value, with one
// placeholder per column, or a condition that matches nothing if a column is not a plain column name.
func SqlLikeAny(cols ...string) string {
	conds := make([]string, len(cols))

	for i, col := range cols {
		if SqlColumn(col) == "" {
			return sqlNoMatch(len(cols))
		}

		conds[i] = SqlLikeCond(col)
	}

	if len(conds) == 0 {
		return sqlNoMatch(0)
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}

// SqlLikeExpr returns an SQL expression that escapes the LIKE wildcards in the value of col, like
// SqlLike does for a bound value, or NULL, which matches nothing, if col is not a plain column name.
// The escape character is replaced innermost, so the ones the outer calls add are not escaped again.
func SqlLikeExpr(col string) string {
	if SqlColumn(col) == "" {
		return "NULL"
	}

	e := SqlLikeEscape

	return fmt.Sprintf("REPLACE(REPLACE(REPLACE(%s, '%s', '%s'), '%%', '%s%%'), '_', '%s_')", col, e, e+e, e, e)
}

// SqlPrefixCond returns a condition matching the values of col that start with a prefix, compared
// case-sensitively on every driver, or a condition that matches nothing if col is not a plain column
// name. Bind the arguments SqlPrefixArgs returns; the LIKE clause keeps an index usable for the lookup.
func SqlPrefixCond(col string) string {
	if SqlColumn(col) == "" {
		return sqlNoMatch(3)
	}

	return fmt.Sprintf("(%s AND SUBSTR(%s, 1, LENGTH(?)) = ?)", SqlLikeCond(col), col)
}

// SqlPrefixArgs returns the arguments of a SqlPrefixCond condition for the given prefix.
func SqlPrefixArgs(prefix string) []any {
	return []any{SqlLike(prefix) + "%", prefix, prefix}
}
