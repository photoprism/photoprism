//go:build ignore

// This generates dialect_mysql.go and dialect_sqlite3.go from the SQL files by running "go generate".
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/fs"
)

// generate writes the Go source that declares the migrations of a dialect.
func generate(name string) {
	dialect := strings.ToLower(name)

	migrations, err := migrate.DialectSource("./"+dialect, dialect)

	if err != nil {
		panic(err)
	}

	code, err := migrate.DialectCode(name, migrations)

	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(fmt.Sprintf("dialect_%s.go", dialect), code, fs.ModeFile); err != nil {
		panic(err)
	}

	fmt.Printf("generated %s with %d migrations\n", dialect, len(migrations))
}

func main() {
	generate("MySQL")
	generate("SQLite3")
}
