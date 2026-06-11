package entity

import (
	"database/sql"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/maruel/natural"
	"github.com/mattn/go-sqlite3"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// dbConn is the global gorm.DB connection provider.
var dbConn Gorm

// Gorm is a gorm.DB connection provider interface.
type Gorm interface {
	Db() *gorm.DB
}

// DbConn is a gorm.DB connection provider.
type DbConn struct {
	Driver string
	Dsn    string

	once sync.Once
	db   *gorm.DB
}

// Db returns the gorm db connection.
func (g *DbConn) Db() *gorm.DB {
	g.once.Do(g.Open)

	if g.db == nil {
		log.Fatal("migrate: database not connected")
	}

	return g.db
}

// Open creates a new gorm db connection.
func (g *DbConn) Open() {
	gDriver := g.Driver
	if g.Driver == dsn.DriverSQLite3 {
		if !slices.Contains(sql.Drivers(), "sqlite_natural_sort") {
			sql.Register("sqlite_natural_sort", &sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					return conn.RegisterCollation("CUSTOM_NATURAL_SORT", natural.Compare)
				},
			})
		}
		gDriver = "sqlite_natural_sort"
		sqliteDialect, ok := gorm.GetDialect("sqlite3")
		if !ok {
			log.Fatal("Unable to get gorm sqlite3 dialect")
		}
		gorm.RegisterDialect("sqlite_natural_sort", sqliteDialect)
	}

	db, err := gorm.Open(gDriver, g.Dsn)

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			fmt.Printf("gorm.Open(%s, %s) %d\n", gDriver, g.Dsn, i)
			db, err = gorm.Open(gDriver, g.Dsn)

			if db != nil && err == nil {
				break
			} else {
				time.Sleep(5 * time.Second)
			}
		}

		if err != nil || db == nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	}

	db.LogMode(false)
	db.SetLogger(log)
	db.DB().SetMaxIdleConns(4)
	db.DB().SetMaxOpenConns(256)

	g.db = db
}

// Close closes the gorm db connection.
func (g *DbConn) Close() {
	if g.db != nil {
		if err := g.db.Close(); err != nil {
			log.Fatal(err)
		}

		g.db = nil
	}
}

// IsDialect returns true if the given sql dialect is used.
func IsDialect(name string) bool {
	return name == Db().Dialect().GetName()
}

// DbDialect returns the sql dialect name.
func DbDialect() string {
	return Db().Dialect().GetName()
}

// SetDbProvider sets the Gorm database connection provider.
func SetDbProvider(conn Gorm) {
	dbConn = conn
}

// HasDbProvider returns true if a db provider exists.
func HasDbProvider() bool {
	return dbConn != nil
}
