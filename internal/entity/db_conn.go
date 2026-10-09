package entity

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	// Registers the database drivers GORM opens connections with.
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

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
func (g *DbConn) Close() {
	if g.db != nil {
		if err := g.db.Close(); err != nil {
			log.Fatal(err)
		}

		g.db = nil
	}
}

// IsDialect returns true if the database uses the SQL dialect of the given dialect or driver name.
func IsDialect(name string) bool {
	dialect := dsn.DialectFromDriver(name)

	return dialect != "" && dialect == DbDialect()
}

// DbDialect returns the SQL dialect of the database as one of the dsn.Dialect constants, or the name
// the ORM reports if the dialect is not supported.
func DbDialect() string {
	name := Db().Dialect().GetName()

	if dialect := dsn.DialectFromDriver(name); dialect != "" {
		return dialect
	}

	return name
}

// BatchSize returns the maximum query parameter number based on the current sql database dialect.
func BatchSize() int {
	switch DbDialect() {
	case dsn.DialectSQLite:
		return 333
	default:
		return 1000
	}
}

// SetDbProvider sets the Gorm database connection provider.
func SetDbProvider(conn Gorm) {
	dbConn = conn
}

// HasDbProvider returns true if a db provider exists.
func HasDbProvider() bool {
	return dbConn != nil
}
