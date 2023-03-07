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
		return gorm.Expr("RAND()")
	case SQLite3:
		return gorm.Expr("RANDOM()")
	default:
		return gorm.Expr("RAND()")
	}
}
