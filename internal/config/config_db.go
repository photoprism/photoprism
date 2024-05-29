package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/migrate"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SQL Databases.
// TODO: PostgreSQL support requires upgrading GORM, so generic column data types can be used.
const (
	MySQL    = "mysql"
	MariaDB  = "mariadb"
	Postgres = "postgres"
	SQLite3  = "sqlite3"
)

// SQLite default DSNs.
const (
	SQLiteTestDB    = ".test.db"
	SQLiteMemoryDSN = ":memory:"
)

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	switch strings.ToLower(c.options.DatabaseDriver) {
	case MySQL, MariaDB:
		c.options.DatabaseDriver = MySQL
	case SQLite3, "sqlite", "sqllite", "test", "file", "":
		c.options.DatabaseDriver = SQLite3
	case "tidb":
		log.Warnf("config: database driver 'tidb' is deprecated, using sqlite")
		c.options.DatabaseDriver = SQLite3
		c.options.DatabaseDsn = ""
	default:
		log.Warnf("config: unsupported database driver %s, using sqlite", c.options.DatabaseDriver)
		c.options.DatabaseDriver = SQLite3
		c.options.DatabaseDsn = ""
	}

	return c.options.DatabaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	if c.options.DatabaseDsn == "" {
		switch c.DatabaseDriver() {
		case MySQL, MariaDB:
			databaseServer := c.DatabaseServer()

			// Connect via Unix Domain Socket?
			if strings.HasPrefix(databaseServer, "/") {
				log.Debugf("mariadb: connecting via Unix domain socket")
				databaseServer = fmt.Sprintf("unix(%s)", databaseServer)
			} else {
				databaseServer = fmt.Sprintf("tcp(%s)", databaseServer)
			}

			return fmt.Sprintf(
				"%s:%s@%s/%s?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=%ds",
				c.DatabaseUser(),
				c.DatabasePassword(),
				databaseServer,
				c.DatabaseName(),
				c.DatabaseTimeout(),
			)
		case Postgres:
			return fmt.Sprintf(
				"user=%s password=%s dbname=%s host=%s port=%d connect_timeout=%d sslmode=disable TimeZone=UTC",
				c.DatabaseUser(),
				c.DatabasePassword(),
				c.DatabaseName(),
				c.DatabaseHost(),
				c.DatabasePort(),
				c.DatabaseTimeout(),
			)
		case SQLite3:
			return filepath.Join(c.StoragePath(), "index.db?_busy_timeout=5000")
		default:
			log.Errorf("config: empty database dsn")
			return ""
		}
	}

	return c.options.DatabaseDsn
}

// DatabaseFile returns the filename part of a sqlite database DSN.
func (c *Config) DatabaseFile() string {
	fileName, _, _ := strings.Cut(strings.TrimPrefix(c.DatabaseDsn(), "file:"), "?")
	return fileName
}

// ParseDatabaseDsn parses the database dsn and extracts user, password, database server, and name.
func (c *Config) ParseDatabaseDsn() {
	if c.options.DatabaseDsn == "" || c.options.DatabaseServer != "" {
		return
	}

	d := NewDSN(c.options.DatabaseDsn)

	c.options.DatabaseName = d.Name
	c.options.DatabaseServer = d.Server
	c.options.DatabaseUser = d.User
	c.options.DatabasePassword = d.Password
}

// DatabaseServer the database server.
func (c *Config) DatabaseServer() string {
	c.ParseDatabaseDsn()

	if c.DatabaseDriver() == SQLite3 {
		return ""
	} else if c.options.DatabaseServer == "" {
		return "localhost"
	}

	return c.options.DatabaseServer
}

// DatabaseHost the database server host.
func (c *Config) DatabaseHost() string {
	if c.DatabaseDriver() == SQLite3 {
		return ""
	}

	if s := strings.Split(c.DatabaseServer(), ":"); len(s) > 0 {
		return s[0]
	}

	return c.options.DatabaseServer
}

// DatabasePort the database server port.
func (c *Config) DatabasePort() int {
	const defaultPort = 3306

	if server := c.DatabaseServer(); server == "" {
		return 0
	} else if s := strings.Split(server, ":"); len(s) != 2 {
		return defaultPort
	} else if port, err := strconv.Atoi(s[1]); err != nil {
		return defaultPort
	} else if port < 1 || port > 65535 {
		return defaultPort
	} else {
		return port
	}
}

// DatabasePortString the database server port as string.
func (c *Config) DatabasePortString() string {
	if c.DatabaseDriver() == SQLite3 {
		return ""
	}

	return strconv.Itoa(c.DatabasePort())
}

// DatabaseName the database schema name.
func (c *Config) DatabaseName() string {
	c.ParseDatabaseDsn()

	if c.DatabaseDriver() == SQLite3 {
		return c.DatabaseDsn()
	} else if c.options.DatabaseName == "" {
		return "photoprism"
	}

	return c.options.DatabaseName
}

// DatabaseUser returns the database user name.
func (c *Config) DatabaseUser() string {
	if c.DatabaseDriver() == SQLite3 {
		return ""
	}

	c.ParseDatabaseDsn()

	if c.options.DatabaseUser == "" {
		return "photoprism"
	}

	return c.options.DatabaseUser
}

// DatabasePassword returns the database user password.
func (c *Config) DatabasePassword() string {
	if c.DatabaseDriver() == SQLite3 {
		return ""
	}

	c.ParseDatabaseDsn()

	return c.options.DatabasePassword
}

// DatabaseTimeout returns the TCP timeout in seconds for establishing a database connection:
// - https://github.com/photoprism/photoprism/issues/4059#issuecomment-1989119004
// - https://github.com/go-sql-driver/mysql/blob/master/README.md#timeout
func (c *Config) DatabaseTimeout() int {
	// Ensure that the timeout is between 1 and a maximum
	// of 60 seconds, with a default of 15 seconds.
	if c.options.DatabaseTimeout <= 0 {
		return 15
	} else if c.options.DatabaseTimeout > 60 {
		return 60
	}

	return c.options.DatabaseTimeout
}

// DatabaseConns returns the maximum number of open connections to the database.
func (c *Config) DatabaseConns() int {
	limit := c.options.DatabaseConns

	if limit <= 0 {
		limit = (runtime.NumCPU() * 2) + 16
	}

	if limit > 1024 {
		limit = 1024
	}

	return limit
}

// DatabaseConnsIdle returns the maximum number of idle connections to the database (equal or less than open).
func (c *Config) DatabaseConnsIdle() int {
	limit := c.options.DatabaseConnsIdle

	if limit <= 0 {
		limit = runtime.NumCPU() + 8
	}

	if limit > c.DatabaseConns() {
		limit = c.DatabaseConns()
	}

	return limit
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

// SetDbOptions sets the database collation to unicode if supported.
func (c *Config) SetDbOptions() {
	switch c.DatabaseDriver() {
	case MySQL, MariaDB:
		c.Db().Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	case Postgres:
		// Ignore for now.
	case SQLite3:
		// Not required as unicode is default.
	}
}

// RegisterDb sets the database options and connection provider.
func (c *Config) RegisterDb() {
	c.SetDbOptions()
	entity.SetDbProvider(c)
}

// InitDb initializes the database without running previously failed migrations.
func (c *Config) InitDb() {
	c.RegisterDb()
	c.MigrateDb(false, nil)
}

// MigrateDb initializes the database and migrates the schema if needed.
func (c *Config) MigrateDb(runFailed bool, ids []string) {
	entity.Admin.UserName = c.AdminUser()

	// Only migrate once automatically per version.
	version := migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.Edition()))
	entity.InitDb(migrate.Opt(version.NeedsMigration(), runFailed, ids))
	if err := version.Migrated(c.Db()); err != nil {
		log.Warnf("config: %s (migrate)", err)
	}

	// Init admin account?
	if c.AdminPassword() == "" {
		log.Warnf("config: password required to initialize %s account", clean.LogQuote(c.AdminUser()))
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword())
	}

	go entity.Error{}.LogEvents()
}

// InitTestDb drops all tables in the currently configured database and re-creates them.
func (c *Config) InitTestDb() {
	entity.ResetTestFixtures()

	if c.AdminPassword() == "" {
		// Do nothing.
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword())
	}

	go entity.Error{}.LogEvents()
}

// checkDb checks the database server version.
func (c *Config) checkDb(db *gorm.DB) error {
	switch c.DatabaseDriver() {
	case MySQL:
		type Res struct {
			Value string `gorm:"column:Value;"`
		}
		var res Res
		if err := db.Raw("SHOW VARIABLES LIKE 'innodb_version'").Scan(&res).Error; err != nil {
			return nil
		} else if v := strings.Split(res.Value, "."); len(v) < 3 {
			log.Warnf("config: unknown database server version")
		} else if major := txt.UInt(v[0]); major < 10 {
			return fmt.Errorf("config: MySQL %s is not supported, see https://docs.photoprism.app/getting-started/#databases", res.Value)
		} else if sub := txt.UInt(v[1]); sub < 5 || sub == 5 && txt.UInt(v[2]) < 12 {
			return fmt.Errorf("config: MariaDB %s is not supported, see https://docs.photoprism.app/getting-started/#databases", res.Value)
		}
	}

	return nil
}

// connectDb establishes a database connection.
func (c *Config) connectDb() error {
	// Make sure this is not running twice.
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	// Get database driver and data source name.
	dbDriver := c.DatabaseDriver()
	dbDsn := c.DatabaseDsn()

	if dbDriver == "" {
		return errors.New("config: database driver not specified")
	}

	if dbDsn == "" {
		return errors.New("config: database DSN not specified")
	}

	// Open database connection.
	db, err := gorm.Open(dbDriver, dbDsn)
	if err != nil || db == nil {
		log.Infof("config: waiting for the database to become available")

		for i := 1; i <= 12; i++ {
			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}

		if err != nil || db == nil {
			return err
		}
	}

	// Configure database logging.
	db.LogMode(false)
	db.SetLogger(log)

	// Set database connection parameters.
	db.DB().SetMaxOpenConns(c.DatabaseConns())
	db.DB().SetMaxIdleConns(c.DatabaseConnsIdle())
	db.DB().SetConnMaxLifetime(time.Hour)

	// Check database server version.
	if err = c.checkDb(db); err != nil {
		if c.Unsafe() {
			log.Error(err)
		} else {
			return err
		}
	}

	// Ok.
	c.db = db

	return nil
}

// ImportSQL imports a file to the currently configured database.
func (c *Config) ImportSQL(filename string) {
	contents, err := os.ReadFile(filename)

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
