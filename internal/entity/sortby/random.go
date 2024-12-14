package sortby

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MySQL    = "mysql"
	SQLite3  = "sqlite"
	Postgres = "postgres"
)

// RandomExpr returns the name of the random function depending on the SQL dialect.
func RandomExpr(dialect gorm.Dialector) clause.Expr {
	switch dialect.Name() {
	case MySQL:
		// A seed integer can be passed as an argument, e.g. "RAND(2342)", to generate
		// reproducible pseudo-random values, see https://mariadb.com/kb/en/rand/.
		return gorm.Expr("RAND()")
	case SQLite3:
		// SQLite does not support specifying a seed to generate a deterministic sequence
		// of pseudo-random values, see https://www.sqlite.org/lang_corefunc.html#random.
		return gorm.Expr("RANDOM()")
	case Postgres:
		// Postgres does not support specifying a seed to generate a deterministic sequence
		// of pseudo-random values, see https://www.postgresql.org/docs/8.2/functions-math.html.
		// random value in the range 0.0 <= x < 1.0
		return gorm.Expr("RANDOM()")
	default:
		return gorm.Expr("RAND()")
	}
}
