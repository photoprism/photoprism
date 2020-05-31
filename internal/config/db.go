package config

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
)

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	switch strings.ToLower(c.params.DatabaseDriver) {
	case MySQL, "mariadb":
		c.params.DatabaseDriver = MySQL
	case SQLite, "sqlite", "sqllite", "test", "file", "":
		c.params.DatabaseDriver = SQLite
	case "tidb":
		log.Warnf("config: database driver 'tidb' is deprecated, using sqlite")
		c.params.DatabaseDriver = SQLite
		c.params.DatabaseDsn = ""
	default:
		log.Warnf("config: unsupported database driver %s, using sqlite", c.params.DatabaseDriver)
		c.params.DatabaseDriver = SQLite
		c.params.DatabaseDsn = ""
	}

	return c.params.DatabaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	if c.params.DatabaseDsn == "" {
		switch c.DatabaseDriver() {
		case MySQL:
			return "photoprism:photoprism@tcp(photoprism-db:3306)/photoprism?parseTime=true"
		case SQLite:
			return filepath.Join(c.StoragePath(), "index.db")
		default:
			log.Errorf("config: empty database dsn")
			return ""
		}
	}

	return c.params.DatabaseDsn
}

// DatabaseConns sets the maximum number of open connections to the database.
func (c *Config) DatabaseConns() int {
	if c.params.DatabaseConns > 1024 || c.params.DatabaseConns < 0 {
		return 0
	}

	return c.params.DatabaseConns
}

// Db returns the db connection.
func (c *Config) Db() *gorm.DB {
	if c.db == nil {
		log.Fatal("config: database not connected")
	}

	return c.db
}

// CloseDb closes the db connection (if any).
func (c *Config) CloseDb() error {
	if c.db != nil {
		if err := c.db.Close(); err == nil {
			c.db = nil
		} else {
			return err
		}
	}

	return nil
}

// InitDb will initialize the database connection and schema.
func (c *Config) InitDb() {
	entity.SetDbProvider(c)
	entity.MigrateDb()
	go entity.SaveErrorMessages()
}

// InitTestDb drops all tables in the currently configured database and re-creates them.
func (c *Config) InitTestDb() {
	entity.SetDbProvider(c)
	entity.ResetTestFixtures()
	go entity.SaveErrorMessages()
}

// connectToDatabase establishes a database connection.
// When used with the internal driver, it may create a new database server instance.
// It tries to do this 12 times with a 5 second sleep interval in between.
func (c *Config) connectToDatabase(ctx context.Context) error {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	dbDriver := c.DatabaseDriver()
	dbDsn := c.DatabaseDsn()

	if dbDriver == "" {
		return errors.New("config: database driver not specified")
	}

	if dbDsn == "" {
		return errors.New("config: database DSN not specified")
	}

	db, err := gorm.Open(dbDriver, dbDsn)
	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}

		if err != nil || db == nil {
			log.Fatal(err)
		}
	}

	db.LogMode(false)
	db.SetLogger(log)

	if runtime.NumCPU() > 4 {
		db.DB().SetMaxIdleConns(runtime.NumCPU())
	} else {
		db.DB().SetMaxIdleConns(4)
	}

	db.DB().SetConnMaxLifetime(time.Minute)
	db.DB().SetMaxOpenConns(c.DatabaseConns())

	c.db = db
	return err
}

// ImportSQL imports a file to the currently configured database.
func (c *Config) ImportSQL(filename string) {
	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Error(err)
		return
	}

	statements := strings.Split(string(contents), ";\n")
	q := c.Db().Unscoped()

	for _, stmt := range statements {
		// Skip empty lines and comments
		if len(stmt) < 3 || stmt[0] == '#' || stmt[0] == ';' {
			continue
		}

		var result struct{}

		q.Raw(stmt).Scan(&result)
	}
}
