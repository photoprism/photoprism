package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratedDialects checks that the generated dialect files match the SQL files they are generated from.
func TestGeneratedDialects(t *testing.T) {
	for name, dialect := range map[string]string{"MySQL": "mysql", "SQLite3": "sqlite3"} {
		t.Run(name, func(t *testing.T) {
			migrations, err := DialectSource(dialect, dialect)
			require.NoError(t, err)

			code, err := DialectCode(name, migrations)
			require.NoError(t, err)

			generated, err := os.ReadFile("dialect_" + dialect + ".go") //nolint:gosec // G304: fixed package file names
			require.NoError(t, err)

			if string(code) != string(generated) {
				t.Fatalf("dialect_%s.go does not match the SQL files in %s/: add or change migrations there and run \"go generate ./internal/entity/migrate\" instead of editing the generated file", dialect, dialect)
			}
		})
	}
	t.Run("Declared", func(t *testing.T) {
		mysql, err := DialectSource("mysql", "mysql")
		require.NoError(t, err)
		sqlite, err := DialectSource("sqlite3", "sqlite3")
		require.NoError(t, err)

		// declaredIDs returns the migration IDs in order, to keep a failure readable.
		declaredIDs := func(m Migrations) (ids []string) {
			for i := range m {
				ids = append(ids, m[i].ID+"."+m[i].Stage)
			}
			return ids
		}

		assert.Equal(t, declaredIDs(mysql), declaredIDs(DialectMySQL))
		assert.Equal(t, declaredIDs(sqlite), declaredIDs(DialectSQLite3))
		assert.True(t, assert.ObjectsAreEqual(mysql, DialectMySQL), "DialectMySQL does not match the SQL files in mysql/")
		assert.True(t, assert.ObjectsAreEqual(sqlite, DialectSQLite3), "DialectSQLite3 does not match the SQL files in sqlite3/")
	})
}

func TestDialectSource(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000001.sql"), []byte("UPDATE a SET b = 1;\nUPDATE c SET d = 2\n"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000002.pre.sql"), []byte("ALTER TABLE a RENAME COLUMN b TO c;\n"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000003.post.sql"), []byte("CREATE INDEX x ON a (c);\n"), 0o600))
		require.NoError(t, os.Mkdir(filepath.Join(dir, "skipped"), 0o700))

		migrations, err := DialectSource(dir, "mysql")
		require.NoError(t, err)
		require.Len(t, migrations, 3)
		assert.Equal(t, StagePost, migrations[2].Stage)
		assert.Equal(t, Migration{ID: "20260101-000001", Dialect: "mysql", Stage: StageMain, Statements: []string{"UPDATE a SET b = 1;", "UPDATE c SET d = 2;"}}, migrations[0])
		assert.Equal(t, Migration{ID: "20260101-000002", Dialect: "mysql", Stage: "pre", Statements: []string{"ALTER TABLE a RENAME COLUMN b TO c;"}}, migrations[1])
	})
	t.Run("InvalidFilename", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000001.txt"), []byte("UPDATE a SET b = 1;"), 0o600))

		_, err := DialectSource(dir, "mysql")
		require.EqualError(t, err, "invalid migration filename 20260101-000001.txt")
	})
	t.Run("InvalidNames", func(t *testing.T) {
		for _, name := range []string{"20260101-000001.pre.txt", "20260101-000001.pre.x.sql", "20260101-000001..sql", ".sql", "20260101-000001.sql~", "20260101-000001.prost.sql", "20260101-000001.main.sql"} {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("UPDATE a SET b = 1;"), 0o600))

			_, err := DialectSource(dir, "mysql")
			require.EqualError(t, err, "invalid migration filename "+name, name)
		}
	})
	t.Run("DuplicateID", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000001.sql"), []byte("UPDATE a SET b = 1;"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000001.post.sql"), []byte("UPDATE a SET b = 2;"), 0o600))

		_, err := DialectSource(dir, "mysql")
		require.EqualError(t, err, "migration id 20260101-000001 is used by 20260101-000001.post.sql and 20260101-000001.sql")
	})
	t.Run("Empty", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "20260101-000001.sql"), []byte(" \n"), 0o600))

		_, err := DialectSource(dir, "mysql")
		require.EqualError(t, err, "migration 20260101-000001.sql is empty")
	})
	t.Run("MissingDir", func(t *testing.T) {
		_, err := DialectSource(filepath.Join(t.TempDir(), "missing"), "mysql")
		require.Error(t, err)
	})
}

func TestSplitStatements(t *testing.T) {
	t.Run("AddsSemicolon", func(t *testing.T) {
		assert.Equal(t, []string{"A;", "B;"}, splitStatements([]byte("A;\n\nB\n")))
	})
	t.Run("FinalSemicolonWithoutNewline", func(t *testing.T) {
		assert.Equal(t, []string{"A;", "B;"}, splitStatements([]byte("A;\nB;")))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, splitStatements([]byte("\n")))
	})
}

func TestDialectCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		code, err := DialectCode("SQLite3", Migrations{{ID: "20260101-000001", Dialect: "sqlite3", Stage: StageMain, Statements: []string{"UPDATE a SET b = 1;"}}})
		require.NoError(t, err)
		assert.Contains(t, string(code), "// Code generated by go generate ./internal/entity/migrate; DO NOT EDIT.\n\npackage migrate\n")
		assert.Contains(t, string(code), "// DialectSQLite3 lists the migrations that run on SQLite.\nvar DialectSQLite3 = Migrations{")
		assert.Contains(t, string(code), `Statements: []string{"UPDATE a SET b = 1;"},`)
	})
	t.Run("UnknownDialect", func(t *testing.T) {
		_, err := DialectCode("Postgres", nil)
		require.EqualError(t, err, "unknown dialect Postgres")
	})
}
