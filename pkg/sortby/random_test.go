package sortby

import (
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/assert"
)

func TestRandomExpr(t *testing.T) {
	mysql, _ := gorm.GetDialect(MySQL)
	sqlite3, _ := gorm.GetDialect(SQLite3)

	assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(mysql))
	assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(sqlite3))
}
