package sortby

import (
	"strings"
)

// Sort direction strings.
const (
	DirAsc    = "ASC"
	DirDesc   = "DESC"
	NullFirst = "FIRST"
	NullLast  = "LAST"
)

// OrderReplacer replaces "ASC" with "DESC" and "DESC" with "ASC", "FIRST" with "LAST" and "LAST" with "FIRST"
var OrderReplacer = strings.NewReplacer(DirAsc, DirDesc, DirDesc, DirAsc, NullFirst, NullLast, NullLast, NullFirst)

// OrderExpr replaces "ASC" with "DESC" and "DESC" with "ASC", "FIRST" with "LAST" and "LAST" with "FIRST" in the specified query order string if reverse is true.
// First and Last are for PostgreSQL NULL ordering
func OrderExpr(s string, reverse bool, dialect string) string {
	if s == "" {
		return ""
	} else if reverse {
		return DialectOrderByFix(OrderReplacer.Replace(s), dialect)
	}

	return DialectOrderByFix(s, dialect)
}

// OrderAsc returns the expression used for sorting in ascending order.
func OrderAsc(reverse bool) string {
	if reverse {
		return DirDesc
	}

	return DirAsc

}

// OrderDesc returns the expression used for sorting in descending order.
func OrderDesc(reverse bool) string {
	if reverse {
		return DirAsc
	}

	return DirDesc
}
