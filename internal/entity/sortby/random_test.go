package sortby

import (
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/constants"

	"github.com/stretchr/testify/assert"
)

func TestRandomExpr(t *testing.T) {
	mysql, _ := gorm.GetDialect(constants.MySQL)
	sqlite3, _ := gorm.GetDialect(constants.SQLite3)

	assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(mysql))
	assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(sqlite3))
}
