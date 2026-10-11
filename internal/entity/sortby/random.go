package sortby

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// RandomExpr returns the random function of an SQL dialect, given as a dialect or driver name.
func RandomExpr(dialect string) clause.Expr {
	switch dsn.DialectFromDriver(dialect) {
	case dsn.DialectMySQL:
		// A seed integer can be passed as an argument, e.g. "RAND(2342)", to generate
		// reproducible pseudo-random values, see https://mariadb.com/kb/en/rand/.
		return gorm.Expr("RAND()")
	case dsn.DialectSQLite, dsn.DialectPostgreSQL:
		// SQLite does not support specifying a seed to generate a deterministic sequence
		// of pseudo-random values, see https://www.sqlite.org/lang_corefunc.html#random.
		// PostgreSQL uses the same name.
		return gorm.Expr("RANDOM()")
	default:
		return gorm.Expr("RAND()")
	}
}
