package entity

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

const (
	// mysqlDefaultCollation is the collation of new tables in a new database.
	mysqlDefaultCollation = "utf8mb4_unicode_ci"
	// mysqlEngine is the storage engine of new tables.
	mysqlEngine = "InnoDB"
)

// mysqlCollationName matches the utf8mb4 collations that new tables may adopt from existing tables.
var mysqlCollationName = regexp.MustCompile(`^utf8mb4_[a-z0-9_]+$`)

// mysqlTable describes the storage options of an existing MySQL/MariaDB table.
type mysqlTable struct {
	Name      string
	Engine    string
	Collation string
}

// Charset returns the character set of the table collation.
func (t mysqlTable) Charset() string {
	charset, _, _ := strings.Cut(t.Collation, "_")
	return charset
}

// mysqlTables returns the storage options of the tables in the list that exist in the current database.
func mysqlTables(db *gorm.DB, list Tables) (result []mysqlTable, err error) {
	var tables []mysqlTable

	if err = db.Raw("SELECT table_name AS name, COALESCE(engine, '') AS engine, " +
		"COALESCE(table_collation, '') AS collation FROM information_schema.tables " +
		"WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' ORDER BY table_name").
		Scan(&tables).Error; err != nil {
		return nil, err
	}

	// The migrations and versions tables are created before the others with the database default,
	// and are not compared with other tables, so they do not count.
	for _, t := range tables {
		if _, ok := list[t.Name]; ok && !schemaTables[t.Name] {
			result = append(result, t)
		}
	}

	return result, nil
}

// mysqlTableOptions returns the options of new tables with the specified collation.
func mysqlTableOptions(collation string) string {
	return fmt.Sprintf("ENGINE=%s DEFAULT CHARSET=utf8mb4 COLLATE=%s", mysqlEngine, collation)
}

// mysqlTableCollation returns the collation of new tables, which the warning also names as conversion
// target: the most common collation of the tables, preferring the default and then utf8mb4 collations on
// a tie, or the default collation for a new database and if the result is not utf8mb4.
func mysqlTableCollation(tables []mysqlTable) string {
	counts := make(map[string]int)

	for _, t := range tables {
		counts[t.Collation]++
	}

	target, best := mysqlDefaultCollation, counts[mysqlDefaultCollation]

	for collation, n := range counts {
		if n > best || n == best && mysqlPreferCollation(collation, target) {
			target, best = collation, n
		}
	}

	if !mysqlCollationName.MatchString(target) {
		return mysqlDefaultCollation
	}

	return target
}

// mysqlPreferCollation reports whether a collation is preferred over the current target on a tie,
// so that the result does not depend on the order of the tables.
func mysqlPreferCollation(collation, target string) bool {
	switch {
	case target == mysqlDefaultCollation:
		return false
	case collation == mysqlDefaultCollation:
		return true
	case mysqlCollationName.MatchString(collation) != mysqlCollationName.MatchString(target):
		return mysqlCollationName.MatchString(collation)
	default:
		return collation < target
	}
}

// warnMysqlTables logs a warning naming the tables that do not use InnoDB, and the tables to convert
// if a table does not use utf8mb4 or the collations differ, since string comparisons between them fail.
func warnMysqlTables(tables []mysqlTable) {
	var engine, convert []string

	collations := make(map[string]bool)
	charset := true

	for _, t := range tables {
		if !strings.EqualFold(t.Engine, mysqlEngine) {
			engine = append(engine, t.Name)
		}

		if t.Charset() != "utf8mb4" {
			charset = false
		}

		collations[t.Collation] = true
	}

	if len(engine) > 0 {
		log.Warnf("migrate: found tables without InnoDB (%s), run \"ALTER TABLE <table> ENGINE=InnoDB;\" "+
			"for each to convert them", strings.Join(engine, ", "))
	}

	if charset && len(collations) <= 1 {
		return
	}

	target := mysqlTableCollation(tables)

	for _, t := range tables {
		if t.Collation != target {
			convert = append(convert, t.Name)
		}
	}

	log.Warnf("migrate: found tables without a shared utf8mb4 collation (%s), run \"ALTER TABLE <table> "+
		"CONVERT TO CHARACTER SET utf8mb4 COLLATE %s;\" for each to convert them",
		strings.Join(convert, ", "), target)
}
