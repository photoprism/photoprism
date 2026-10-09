package sortby

import (
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// RandomExpr returns the random function of an SQL dialect, given as a dialect or driver name.
func RandomExpr(dialect string) *gorm.SqlExpr {
	switch dsn.DialectFromDriver(dialect) {
	case dsn.DialectMySQL:
		// A seed integer can be passed as an argument, e.g. "RAND(2342)", to generate
		// reproducible pseudo-random values, see https://mariadb.com/kb/en/rand/.
		return gorm.Expr("RAND()")
	case dsn.DialectSQLite:
		// SQLite does not support specifying a seed to generate a deterministic sequence
		// of pseudo-random values, see https://www.sqlite.org/lang_corefunc.html#random.
		return gorm.Expr("RANDOM()")
	default:
		return gorm.Expr("RAND()")
	}
}
