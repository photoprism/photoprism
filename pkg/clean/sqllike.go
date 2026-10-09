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
func sqlNoMatch(postgres bool, n int) string {
	if postgres {
		return "(1 = 0" + strings.Repeat(" AND CAST(? AS VARCHAR) IS NULL", n) + ")"
	}
	return "(1 = 0" + strings.Repeat(" AND ? IS NULL", n) + ")"
}

// SqlLikeCond returns "col LIKE ? ESCAPE '!'" (or postgres equivalents), or a condition that matches nothing if col is not a
// plain column name. The drivers disagree on the default escape character, so a SqlLike value needs
// this ESCAPE clause.
func SqlLikeCond(postgres bool, binary bool, col string) string {
	if SqlColumn(col) == "" {
		return sqlNoMatch(postgres, 1)
	}

	switch {
	case postgres && binary:
		return fmt.Sprintf("convert_from(%s, 'UTF8') ILIKE ? ESCAPE '%s'", col, SqlLikeEscape)
	case postgres:
		return fmt.Sprintf("%s ILIKE ? ESCAPE '%s'", col, SqlLikeEscape)
	default:
		return fmt.Sprintf("%s LIKE ? ESCAPE '%s'", col, SqlLikeEscape)
	}
}

// SqlLikeAny returns a condition that matches if any of the columns is LIKE the bound value, with one
// placeholder per column, or a condition that matches nothing if a column is not a plain column name.
// DO NOT USE A MIX OF BINARY AND STRING COLUMNS!!!
func SqlLikeAny(postgres bool, binary bool, cols ...string) string {
	conds := make([]string, len(cols))

	for i, col := range cols {
		if SqlColumn(col) == "" {
			return sqlNoMatch(postgres, len(cols))
		}

		conds[i] = SqlLikeCond(postgres, binary, col)
	}

	if len(conds) == 0 {
		return sqlNoMatch(postgres, 0)
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}

// SqlLikeExpr returns an SQL expression that escapes the LIKE wildcards in the value of col, like
// SqlLike does for a bound value, or NULL, which matches nothing, if col is not a plain column name.
// The escape character is replaced innermost, so the ones the outer calls add are not escaped again.
func SqlLikeExpr(postgres bool, binary bool, col string) string {
	if SqlColumn(col) == "" {
		return "NULL"
	}

	e := SqlLikeEscape

	if postgres && binary {
		return fmt.Sprintf("REPLACE(REPLACE(REPLACE(convert_from(%s, 'UTF8'), '%s', '%s'), '%%', '%s%%'), '_', '%s_')", col, e, e+e, e, e)
	}

	return fmt.Sprintf("REPLACE(REPLACE(REPLACE(%s, '%s', '%s'), '%%', '%s%%'), '_', '%s_')", col, e, e+e, e, e)
}

// SqlPrefixCond returns a condition matching the values of col that start with a prefix, compared
// case-sensitively on every driver, or a condition that matches nothing if col is not a plain column
// name. Bind the arguments SqlPrefixArgs returns; the LIKE clause keeps an index usable for the lookup.
func SqlPrefixCond(postgres bool, binary bool, col string) string {
	if SqlColumn(col) == "" {
		return sqlNoMatch(postgres, 3)
	}

	return fmt.Sprintf("(%s AND SUBSTR(%s, 1, LENGTH(?)) = ?)", SqlLikeCond(postgres, binary, col), col)
}

// SqlPrefixArgs returns the arguments of a SqlPrefixCond condition for the given prefix.
func SqlPrefixArgs(prefix string) []any {
	return []any{SqlLike(prefix) + "%", prefix, prefix}
}
