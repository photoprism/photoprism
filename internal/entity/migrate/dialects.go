package migrate

import (
	"sync"

	"github.com/photoprism/photoprism/pkg/constants"
)

var Dialects = map[string]Migrations{
	constants.MySQL:   DialectMySQL,
	constants.SQLite3: DialectSQLite3,
}

var once = map[string]*sync.Once{
	constants.MySQL:   {},
	constants.SQLite3: {},
}
