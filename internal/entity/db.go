package entity

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Database drivers (sql dialects).
const (
	MySQL   = "mysql"
	SQLite3 = "sqlite3"
)

var dbProvider DbProvider

type DbProvider interface {
	Db() *gorm.DB
}

// IsDialect returns true if the given sql dialect is used.
func IsDialect(name string) bool {
	return name == Db().Dialect().GetName()
}

// DbDialect returns the sql dialect name.
func DbDialect() string {
	return Db().Dialect().GetName()
}

// SetDbProvider sets the provider to get a gorm db connection.
func SetDbProvider(provider DbProvider) {
	dbProvider = provider
}

// HasDbProvider returns true if a db provider exists.
func HasDbProvider() bool {
	return dbProvider != nil
}

// Db returns a database connection.
func Db() *gorm.DB {
	return dbProvider.Db()
}

// UnscopedDb returns an unscoped database connection.
func UnscopedDb() *gorm.DB {
	return Db().Unscoped()
}

type Gorm struct {
	Driver string
	Dsn    string

	once sync.Once
	db   *gorm.DB
}

// Db returns the gorm db connection.
func (g *Gorm) Db() *gorm.DB {
	g.once.Do(g.Connect)

	if g.db == nil {
		log.Fatal("entity: database not connected")
	}

	return g.db
}

// Connect creates a new gorm db connection.
func (g *Gorm) Connect() {
	db, err := gorm.Open(g.Driver, g.Dsn)

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			fmt.Printf("gorm.Open(%s, %s) %d\n", g.Driver, g.Dsn, i)
			db, err = gorm.Open(g.Driver, g.Dsn)

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
func (g *Gorm) Close() {
	if g.db != nil {
		if err := g.db.Close(); err != nil {
			log.Fatal(err)
		}

		g.db = nil
	}
}
