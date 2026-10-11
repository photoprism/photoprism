package search

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// OrderExpr replaces "ASC" with "DESC" and "DESC" with "ASC" in the specified query order string if reverse is true.
func OrderExpr(s string, reverse bool) string {
	return sortby.OrderExpr(s, reverse)
}

// OrderExprNulls returns OrderExpr for an order string with columns that may be NULL, which sorts NULL as the
// smallest value on PostgreSQL too, as on the other databases.
func OrderExprNulls(s string, reverse bool) string {
	s = sortby.OrderExpr(s, reverse)

	if s != "" && entity.DbDialect() == dsn.DialectPostgreSQL {
		return sortby.NullsOrder(s)
	}

	return s
}
