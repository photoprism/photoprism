package migrate

import (
	"sync"

	"github.com/photoprism/photoprism/pkg/enum"
)

var Dialects = map[string]Migrations{
	enum.MySQL:   DialectMySQL,
	enum.SQLite3: DialectSQLite3,
}

var once = map[string]*sync.Once{
	enum.MySQL:   {},
	enum.SQLite3: {},
}
