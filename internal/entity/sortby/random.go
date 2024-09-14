package sortby

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	MySQL   = "mysql"
	SQLite3 = "sqlite3"
)

// RandomExpr returns the name of the random function depending on the SQL dialect.
func RandomExpr(dialect gorm.Dialect) *gorm.SqlExpr {
	switch dialect.GetName() {
	case MySQL:
		// A seed integer can be passed as an argument, e.g. "RAND(2342)", to generate
		// reproducible pseudo-random values, see https://mariadb.com/kb/en/rand/.
		return gorm.Expr("RAND()")
	case SQLite3:
		// SQLite does not support specifying a seed to generate a deterministic sequence
		// of pseudo-random values, see https://www.sqlite.org/lang_corefunc.html#random.
		return gorm.Expr("RANDOM()")
	default:
		return gorm.Expr("RAND()")
	}
}
