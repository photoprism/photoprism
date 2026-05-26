package sortby

import (
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestRandomExpr(t *testing.T) {
	mysql, _ := gorm.GetDialect(dsn.DriverMySQL)
	sqlite3, _ := gorm.GetDialect(dsn.DriverSQLite3)

	assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(mysql))
	assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(sqlite3))
}
