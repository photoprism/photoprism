package config

import (
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/tidb"
)

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	if c.config.DatabaseDriver == "" {
		return DbTiDB
	}

	return c.config.DatabaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	if c.config.DatabaseDsn == "" {
		return "root:photoprism@tcp(localhost:4000)/photoprism?parseTime=true"
	}

	return c.config.DatabaseDsn
}

// Db returns the db connection.
func (c *Config) Db() *gorm.DB {
	if c.db == nil {
		log.Fatal("config: database not initialised")
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

// MigrateDb will start a migration process.
func (c *Config) MigrateDb() {
	db := c.Db()

	db.AutoMigrate(
		&entity.File{},
		&entity.Photo{},
		&entity.Event{},
		&entity.Place{},
		&entity.Location{},
		&entity.Camera{},
		&entity.Lens{},
		&entity.Country{},
		&entity.Share{},

		&entity.Album{},
		&entity.PhotoAlbum{},
		&entity.Label{},
		&entity.Category{},
		&entity.PhotoLabel{},
		&entity.Keyword{},
		&entity.PhotoKeyword{},
	)

	entity.CreateUnknownPlace(db)
	entity.CreateUnknownCountry(db)
}

// connectToDatabase establishes a database connection.
// When used with the internal driver, it may create a new database server instance.
// It tries to do this 12 times with a 5 second sleep interval in between.
func (c *Config) connectToDatabase(ctx context.Context) error {
	dbDriver := c.DatabaseDriver()
	dbDsn := c.DatabaseDsn()

	if dbDriver == "" {
		return errors.New("config: database driver not specified")
	}

	if dbDsn == "" {
		return errors.New("config: database DSN not specified")
	}

	isTiDB := false
	initSuccess := false

	if dbDriver == DbTiDB {
		isTiDB = true
		dbDriver = DbMySQL
	}

	db, err := gorm.Open(dbDriver, dbDsn)
	if err != nil || db == nil {
		if isTiDB {
			log.Infof("starting database server at %s:%d\n", c.SqlServerHost(), c.SqlServerPort())

			go tidb.Start(ctx, c.SqlServerPath(), c.SqlServerPort(), c.SqlServerHost(), c.Debug())
		}

		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			if isTiDB && !initSuccess {
				err = tidb.InitDatabase(c.SqlServerPort(), c.SqlServerPassword())

				if err != nil {
					log.Debug(err)
				} else {
					initSuccess = true
				}
			}
		}

		if err != nil || db == nil {
			log.Fatal(err)
		}
	}

	c.db = db
	return err
}

// DropTables drops all tables in the currently configured database (be careful!).
func (c *Config) DropTables() {
	db := c.Db()

	db.DropTableIfExists(
		&entity.File{},
		&entity.Photo{},
		&entity.Event{},
		&entity.Place{},
		&entity.Location{},
		&entity.Camera{},
		&entity.Lens{},
		&entity.Country{},
		&entity.Share{},

		&entity.Album{},
		&entity.PhotoAlbum{},
		&entity.Label{},
		&entity.Category{},
		&entity.PhotoLabel{},
		&entity.Keyword{},
		&entity.PhotoKeyword{},
	)
}

// ImportSQL imports a file to the currently configured database.
func (c *Config) ImportSQL(filename string) {
	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Error(err)
		return
	}

	statements := strings.Split(string(contents), ";\n")

	for _, stmt := range statements {
		// Skip empty lines and comments
		if len(stmt) < 3 || stmt[0] == '#' || stmt[0] == ';' {
			continue
		}

		if _, err := c.Db().CommonDB().Query(stmt); err != nil {
			log.Error(err)
		}
	}
}
