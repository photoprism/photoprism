package sortby

import (
	"strings"
)

// Sort direction strings.
const (
	DirAsc  = "ASC"
	DirDesc = "DESC"
)

// OrderReplacer replaces "ASC" with "DESC" and "DESC" with "ASC"
var OrderReplacer = strings.NewReplacer(DirAsc, DirDesc, DirDesc, DirAsc)

// OrderExpr replaces "ASC" with "DESC" and "DESC" with "ASC" in the specified query order string if reverse is true.
func OrderExpr(s string, reverse bool) string {
	if s == "" {
		return ""
	} else if reverse {
		return OrderReplacer.Replace(s)
	}

	return s
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
