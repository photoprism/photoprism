package sortby

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestRandomExpr(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(dsn.DialectMySQL))
	})
	t.Run("SQLite", func(t *testing.T) {
		assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(dsn.DialectSQLite))
		assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(dsn.DriverSQLite3))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(""))
	})
}
