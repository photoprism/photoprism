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
	if strings.ToLower(c.params.DatabaseDriver) == "mysql" {
		return DriverMysql
	}

	return DriverTidb
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	if c.params.DatabaseDsn == "" {
		return "root:photoprism@tcp(localhost:2343)/photoprism?parseTime=true"
	}

	return c.params.DatabaseDsn
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
	entity.CreateUnknownCamera(db)
	entity.CreateUnknownLens(db)
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

	if dbDriver == DriverTidb {
		isTiDB = true
		dbDriver = DriverMysql
	}

	db, err := gorm.Open(dbDriver, dbDsn)
	if err != nil || db == nil {
		if isTiDB {
			log.Infof("starting database server at %s:%d\n", c.TidbServerHost(), c.TidbServerPort())

			go tidb.Start(ctx, c.TidbServerPath(), c.TidbServerPort(), c.TidbServerHost(), c.Debug())
		}

		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			if isTiDB && !initSuccess {
				err = tidb.InitDatabase(c.TidbServerPort(), c.TidbServerPassword())

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
