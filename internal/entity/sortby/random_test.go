package sortby

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestRandomExpr(t *testing.T) {
	mysql := mysql.New(mysql.Config{})
	testDbTemp := "./testdata/migrate_sqlite3.db"
	dumpName, err := filepath.Abs(testDbTemp)
	_ = os.Remove(dumpName)
	if err != nil {
		t.Fatal(err)
	}
	sqlite := sqlite.Open(dumpName)

	assert.Equal(t, gorm.Expr("RAND()"), RandomExpr(mysql))
	assert.Equal(t, gorm.Expr("RANDOM()"), RandomExpr(sqlite))
}
