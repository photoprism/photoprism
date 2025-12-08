package search

import (
	"github.com/photoprism/photoprism/internal/entity/sortby"
)

// OrderExpr replaces "ASC" with "DESC" and "DESC" with "ASC" in the specified query order string if reverse is true.
func OrderExpr(s string, reverse bool) string {
	return sortby.OrderExpr(s, reverse)
}
