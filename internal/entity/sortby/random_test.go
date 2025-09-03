package sortby

import (
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/enum"

	"github.com/stretchr/testify/assert"
)

func TestRandomExpr(t *testing.T) {
	mysql, _ := gorm.GetDialect(enum.MySQL)
	sqlite3, _ := gorm.GetDialect(enum.SQLite3)

	assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(mysql))
	assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(sqlite3))
}
