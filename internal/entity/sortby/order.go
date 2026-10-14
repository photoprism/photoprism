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

// NullsOrder returns an order string with "NULLS FIRST" after each ascending and "NULLS LAST" after each
// descending term without one, so that NULL sorts as the smallest value, as on MariaDB and SQLite, also on
// databases that sort it as the largest, such as PostgreSQL.
func NullsOrder(s string) string {
	terms := splitOrderTerms(s)

	for i, term := range terms {
		t := strings.TrimSpace(term)
		upper := strings.ToUpper(t)

		switch {
		case t == "", strings.Contains(upper, " NULLS "):
			terms[i] = t
		case strings.HasSuffix(upper, " "+DirDesc):
			terms[i] = t + " NULLS LAST"
		default:
			terms[i] = t + " NULLS FIRST"
		}
	}

	return strings.Join(terms, ", ")
}

// splitOrderTerms splits an order string at the commas that are not enclosed in parentheses.
func splitOrderTerms(s string) (terms []string) {
	depth, start := 0, 0

	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				terms = append(terms, s[start:i])
				start = i + 1
			}
		}
	}

	return append(terms, s[start:])
}
