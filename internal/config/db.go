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
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/tidb"
	"github.com/sirupsen/logrus"
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
		&entity.Account{},
		&entity.File{},
		&entity.FileShare{},
		&entity.FileSync{},
		&entity.Photo{},
		&entity.Description{},
		&entity.Event{},
		&entity.Place{},
		&entity.Location{},
		&entity.Camera{},
		&entity.Lens{},
		&entity.Country{},
		&entity.Album{},
		&entity.PhotoAlbum{},
		&entity.Label{},
		&entity.Category{},
		&entity.PhotoLabel{},
		&entity.Keyword{},
		&entity.PhotoKeyword{},
		&entity.Link{},
	)

	entity.CreateUnknownPlace(db)
	entity.CreateUnknownCountry(db)
}

// DropTables drops all tables in the currently configured database (be careful!).
func (c *Config) DropTables() {
	db := c.Db()

	logLevel := log.Level

	log.SetLevel(logrus.FatalLevel)
	db.SetLogger(log)
	db.LogMode(false)

	db.DropTableIfExists(
		&entity.Account{},
		&entity.File{},
		&entity.FileShare{},
		&entity.FileSync{},
		&entity.Photo{},
		&entity.Description{},
		&entity.Event{},
		&entity.Place{},
		&entity.Location{},
		&entity.Camera{},
		&entity.Lens{},
		&entity.Country{},
		&entity.Album{},
		&entity.PhotoAlbum{},
		&entity.Label{},
		&entity.Category{},
		&entity.PhotoLabel{},
		&entity.Keyword{},
		&entity.PhotoKeyword{},
		&entity.Link{},
	)

	log.SetLevel(logLevel)
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

			go tidb.Start(ctx, c.DatabasePath(), c.SqlServerPort(), c.SqlServerHost(), c.Debug())
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

	db.LogMode(false)
	db.SetLogger(log)

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

		err := q.Raw(stmt).Scan(&result).Error

		if err != nil {
			log.Error(err)
		}
	}
}
