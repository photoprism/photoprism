package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// Tables maps a table name to the entity that defines its schema.
type Tables map[string]any

// schemaTables contains the tables that record the schema migration state.
// Truncate keeps them, as the schema remains migrated when only data is removed.
var schemaTables = map[string]bool{
	migrate.Migration{}.TableName(): true,
	migrate.Version{}.TableName():   true,
}

// Entities contains database entities and their table names.
var Entities = Tables{
	migrate.Migration{}.TableName(): &migrate.Migration{},
	migrate.Version{}.TableName():   &migrate.Version{},
	Error{}.TableName():             &Error{},
	Password{}.TableName():          &Password{},
	Passcode{}.TableName():          &Passcode{},
	User{}.TableName():              &User{},
	UserDetails{}.TableName():       &UserDetails{},
	UserSettings{}.TableName():      &UserSettings{},
	Session{}.TableName():           &Session{},
	Client{}.TableName():            &Client{},
	Service{}.TableName():           &Service{},
	Folder{}.TableName():            &Folder{},
	Duplicate{}.TableName():         &Duplicate{},
	File{}.TableName():              &File{},
	FileShare{}.TableName():         &FileShare{},
	FileSync{}.TableName():          &FileSync{},
	Photo{}.TableName():             &Photo{},
	PhotoUser{}.TableName():         &PhotoUser{},
	Details{}.TableName():           &Details{},
	Place{}.TableName():             &Place{},
	Cell{}.TableName():              &Cell{},
	Camera{}.TableName():            &Camera{},
	Lens{}.TableName():              &Lens{},
	Country{}.TableName():           &Country{},
	Album{}.TableName():             &Album{},
	AlbumUser{}.TableName():         &AlbumUser{},
	PhotoAlbum{}.TableName():        &PhotoAlbum{},
	Label{}.TableName():             &Label{},
	Category{}.TableName():          &Category{},
	PhotoLabel{}.TableName():        &PhotoLabel{},
	Keyword{}.TableName():           &Keyword{},
	PhotoKeyword{}.TableName():      &PhotoKeyword{},
	Link{}.TableName():              &Link{},
	Subject{}.TableName():           &Subject{},
	Face{}.TableName():              &Face{},
	Marker{}.TableName():            &Marker{},
	Reaction{}.TableName():          &Reaction{},
	UserShare{}.TableName():         &UserShare{},
}

// WaitForMigration waits for the database migration to be successful and returns an error otherwise.
func (list Tables) WaitForMigration(db *gorm.DB) error {
	type RowCount struct {
		Count int
	}

	attempts := 100
	for name := range list {
		for i := 0; i <= attempts; i++ {
			count := RowCount{}
			if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", name)).Scan(&count).Error; err == nil {
				log.Tracef("migrate: %s migrated", clean.Log(name))
				break
			} else {
				log.Tracef("migrate: waiting for %s migration (%s)", clean.Log(name), err.Error())
				time.Sleep(100 * time.Millisecond)
			}

			if i == attempts {
				return errors.New("some database tables are missing")
			}
		}
	}

	return nil
}

// Truncate removes all data from tables without dropping them, except for the
// schemaTables that record the migration state.
func (list Tables) Truncate(db *gorm.DB) {
	var name string

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (truncate)", r, name)
		}
	}()

	autoInc := autoIncrementTables(db)

	for name = range list {
		if schemaTables[name] {
			continue
		} else if err := truncateTable(db, name, autoInc[name]); err != nil && err.Error() != "record not found" {
			log.Debugf("migrate: %s in %s", err, clean.Log(name))
		}
	}
}

// truncateTable removes all rows from the table with the specified name, so that
// generated IDs match those in a newly created database.
//
// MySQL/MariaDB delete the rows and restart an auto-increment counter, as TRUNCATE
// is a slower DDL statement there; the restart commits implicitly, so pass db, not
// a transaction. SQLite deletes the rows and keeps its counters; others truncate.
func truncateTable(db *gorm.DB, name string, autoInc bool) error {
	switch dsn.DialectFromDriver(db.Dialector.Name()) {
	case dsn.DialectSQLite:
		return db.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1", name)).Error
	case dsn.DialectMySQL:
		if err := RetryDeadlock("truncate "+name, func() error {
			return db.Exec(fmt.Sprintf("DELETE FROM `%s`", name)).Error
		}); err != nil {
			return err
		} else if autoInc {
			return db.Exec(fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT = 1", name)).Error
		}

		return nil
	}

	if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", name)).Error; err == nil {
		return nil
	}

	return db.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1", name)).Error
}

// autoIncrementTables returns the MySQL/MariaDB tables in the current database
// that have an auto-increment column.
func autoIncrementTables(db *gorm.DB) map[string]bool {
	result := make(map[string]bool)

	if dsn.DialectFromDriver(db.Dialector.Name()) != dsn.DialectMySQL {
		return result
	}

	rows, err := db.Raw("SELECT table_name FROM information_schema.columns WHERE table_schema = DATABASE() AND extra LIKE '%auto_increment%'").Rows()

	if err != nil {
		log.Debugf("migrate: %s (find auto-increment tables)", err)
		return result
	}

	defer rows.Close()

	for rows.Next() {
		var name string

		if err = rows.Scan(&name); err != nil {
			log.Debugf("migrate: %s (scan auto-increment tables)", err)
			continue
		}

		result[name] = true
	}

	if err = rows.Err(); err != nil {
		log.Debugf("migrate: %s (find auto-increment tables)", err)
	}

	return result
}

// Migrate migrates all database tables of registered entities, and returns an error if the schema
// of a table could not be migrated, so the version is not recorded as migrated.
func (list Tables) Migrate(db *gorm.DB, opt migrate.Options) (err error) {
	var name string
	var entity any

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (panic)", r, name)
			err = fmt.Errorf("migrate: %s in %s", r, name)
		}
	}()

	log.Debugf("migrate: running database migrations")

	// Run pre migrations, if any.
	if err := migrate.Run(db, opt.Pre()); err != nil {
		log.Error(err)
	}

	// New MySQL/MariaDB tables get the options of the existing tables, which are checked on every start.
	tableDb := withTableOptions(db, list, true)

	// Run ORM auto migrations.
	if opt.AutoMigrate {
		// Recreate SQLite tables with legacy column types once, after the pre migrations renamed columns.
		if err = migrate.ConvertSQLiteSchema(db, list); err != nil {
			for _, entity = range list {
				logDuplicateKeys(db, entity)
			}

			return err
		}

		// Link the categories table to the label categories relation before tables are created.
		if err = db.SetupJoinTable(&Label{}, "LabelCategories", &Category{}); err != nil {
			return err
		}

		for name, entity = range list {
			if err := tableDb.AutoMigrate(entity); err != nil {
				log.Debugf("migrate: %s (waiting 1s)", err.Error())

				time.Sleep(time.Second)

				if err = tableDb.AutoMigrate(entity); err != nil {
					log.Errorf("migrate: failed migrating %s", clean.Log(name))
					logDuplicateKeys(db, entity)
					panic(err)
				}
			}
		}
	}

	// Run main migrations, if any.
	if err := migrate.Run(db, opt); err != nil {
		log.Error(err)
	}

	// Run post migrations, if any.
	if err := migrate.Run(db, opt.Post()); err != nil {
		log.Error(err)
	}

	return nil
}

// logDuplicateKeys logs the unique indexes of an entity whose columns hold duplicate values, as the
// index cannot be created until they are removed, with the query that lists them.
func logDuplicateKeys(db *gorm.DB, entity any) {
	stmt := &gorm.Statement{DB: db}

	if err := stmt.Parse(entity); err != nil || !db.Migrator().HasTable(entity) {
		return
	}

	for _, idx := range stmt.Schema.ParseIndexes() {
		if idx.Class != "UNIQUE" || len(idx.Fields) == 0 {
			continue
		}

		cols := make([]string, len(idx.Fields))
		cond := make([]string, len(idx.Fields))

		for i, f := range idx.Fields {
			cols[i] = stmt.Quote(f.DBName)
			cond[i] = cols[i] + " IS NOT NULL"
		}

		list := strings.Join(cols, ", ")
		query := fmt.Sprintf("SELECT %s, COUNT(*) FROM %s WHERE %s GROUP BY %s HAVING COUNT(*) > 1",
			list, stmt.Quote(stmt.Table), strings.Join(cond, " AND "), list)

		var count int64

		if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM (%s) d", query)).Scan(&count).Error; err != nil || count == 0 {
			continue
		}

		event.SystemError([]string{"migrate", "cannot create unique index %s on %s with duplicate values (%d), " +
			"run \"%s;\" to list them"}, idx.Name, stmt.Table, count, query)
	}
}

// Drop drops all database tables of registered entities.
func (list Tables) Drop(db *gorm.DB) {
	for _, entity := range list {
		if err := db.Migrator().DropTable(entity); err != nil {
			panic(err)
		}
	}
}

// withTableOptions returns a session that creates MySQL/MariaDB tables with InnoDB, utf8mb4 and the collation
// of the existing tables in the list, and logs a warning for mismatched table options if warn is true.
func withTableOptions(db *gorm.DB, list Tables, warn bool) *gorm.DB {
	if dsn.DialectFromDriver(db.Name()) != dsn.DialectMySQL {
		return db
	}

	tables, err := mysqlTables(db, list)

	if err != nil {
		log.Debugf("migrate: %s (find table options)", err)
	} else if warn {
		warnMysqlTables(tables)
	}

	options := mysqlTableOptions(mysqlTableCollation(tables))

	return db.Session(&gorm.Session{}).Set("gorm:table_options", options).Session(&gorm.Session{})
}

// CreateTable creates the table of an entity, including its indexes.
func CreateTable(db *gorm.DB, entity any) error {
	return withTableOptions(db, Entities, false).Migrator().CreateTable(entity)
}
