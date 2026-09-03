//go:build ignore

// This generates countries.go by running "go generate"
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/photoprism/photoprism/internal/entity/migrate"
)

func gen_migrations(name string) {
	if name == "" {
		return
	}

	dialect := strings.ToLower(name)

	type Migration struct {
		ID         string
		Stage      string
		Dialect    string
		Statements []string
	}

	var migrations []Migration

	// Folder in which migration files are stored.
	folder := "./" + dialect

	// Returns directory entries sorted by filename.
	files, _ := os.ReadDir(folder)

	fmt.Printf("generating %s...", dialect)

	strToStmts := func(b []byte) (result []string) {
		stmts := bytes.Split(b, []byte(";\n"))
		result = make([]string, 0, len(stmts))

		for i := range stmts {
			if s := bytes.TrimSpace(stmts[i]); len(s) > 0 {
				if s[len(s)-1] != ';' {
					s = append(s, ';')
				}

				result = append(result, string(s))
			}
		}

		return result
	}

	// Postgres has functions that can be declared, with $$ as a wrapper around ;\n.
	if strings.ToLower(dialect) == "postgres" {
		strToStmts = func(b []byte) (result []string) {
			runes := []rune(string(b))

			if len(runes) == 0 {
				return nil
			}

			var (
				statements []string
				current    strings.Builder
				inQuotes   bool
				inDollars  bool
				quoteRune  rune
			)

			flush := func() {
				if current.Len() == 0 || len(strings.TrimSpace(current.String())) == 0 {
					return
				}
				statements = append(statements, strings.TrimSpace(current.String()))
				current.Reset()
			}

			for i := 0; i < len(runes); i++ {
				r := runes[i]

				switch {
				case inQuotes && r == '\\':
					if i+1 < len(runes) {
						current.WriteRune(r)
						current.WriteRune(runes[i+1])
						i++
					}
				case r == '\'' || r == '"':
					if inQuotes {
						if i+1 < len(runes) {
							if runes[i+1] == quoteRune {
								current.WriteRune(r)
								current.WriteRune(runes[i+1])
								i++
							} else {
								inQuotes = false
								current.WriteRune(r)
							}
						} else {
							if r == quoteRune {
								inQuotes = false
							}
							current.WriteRune(r)
						}
					} else {
						current.WriteRune(r)
						inQuotes = true
						quoteRune = r
					}
				case r == '$':
					if inQuotes {
						current.WriteRune(r)
					} else {
						if i+1 < len(runes) {
							if inDollars {
								if runes[i+1] == '$' {
									current.WriteRune(r)
									current.WriteRune(r)
									inDollars = false
									i++
								} else {
									current.WriteRune(r)
								}
							} else {
								if runes[i+1] == '$' {
									current.WriteRune(r)
									current.WriteRune(r)
									inDollars = true
									i++
								} else {
									current.WriteRune(r)
								}
							}
						} else {
							current.WriteRune(r)
						}
					}
				case r == ';':
					if inQuotes || inDollars {
						current.WriteRune(r)
					} else {
						current.WriteRune(r)
						flush()
					}
				default:
					current.WriteRune(r)
				}
			}

			flush()

			if len(statements) == 0 {
				return nil
			}
			return statements
		}
	}

	// Read migrations from files.
	for _, file := range files {
		stage := ""
		filePath := filepath.Join(folder, file.Name())
		fileName := strings.SplitN(filepath.Base(file.Name()), ".", 3)

		if file.IsDir() {
			// Skip directory.
			continue
		} else if len(fileName) < 2 || fileName[0] == "" {
			// Invalid filename.
			fmt.Printf("e")
			continue
		} else if fileName[1] != "sql" && fileName[2] != "sql" {
			// Invalid filename.
			fmt.Printf("e")
			continue
		} else if fileName[1] != "sql" {
			// Stage, if any.
			stage = fileName[1]
		} else {
			stage = migrate.StageMain
		}

		// Migration ID.
		id := fileName[0]

		// Extract SQL from file.
		if s, err := os.ReadFile(filePath); err == nil && len(s) > 0 {
			fmt.Printf(".")
			migrations = append(migrations, Migration{ID: id, Stage: stage, Dialect: dialect, Statements: strToStmts(s)})
		} else {
			fmt.Printf("f")
			fmt.Println(err.Error())
		}
	}

	fmt.Printf(" found %d migrations\n", len(migrations))

	// Create source file from migrations.
	f, err := os.Create(fmt.Sprintf("dialect_%s.go", dialect))

	if err != nil {
		panic(err)
	}

	defer f.Close()

	// Render source template.
	migrationsTemplate.Execute(f, struct {
		Name       string
		Migrations []Migration
	}{
		Name:       name,
		Migrations: migrations,
	})
}

func main() {
	gen_migrations("MySQL")
	gen_migrations("SQLite")
	gen_migrations("Postgres")
}

var migrationsTemplate = template.Must(template.New("").Parse(`package migrate

// Generated code, do not edit.

// Dialect{{ print .Name }} is the migrations for the DBMS {{ print .Name }}
var Dialect{{ print .Name }} = Migrations{
{{- range .Migrations }}
	{
		ID:         {{ printf "%q" .ID }},
		Dialect:    {{ printf "%q" .Dialect }},
		Stage:      {{ printf "%q" .Stage }},
		Statements: []string{ {{- range $index, $s := .Statements}}{{if $index}}, {{end}}{{ printf "%q" $s }}{{end -}} },
	},	
{{- end }}
}
`))
