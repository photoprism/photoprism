package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
)

// SQL Databases.
const (
	MySQL    = "mysql"
	MariaDB  = "mariadb"
	Postgres = "postgres"
	SQLite3  = "sqlite"
)

// SQLite default DSNs.
const (
	SQLiteTestDB    = ".test.db"
	SQLiteMemoryDSN = ":memory:"
)

var drivers = map[string]func(string) gorm.Dialector{
	MySQL:    mysql.Open,
	SQLite3:  sqlite.Open,
	Postgres: postgres.Open,
}

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	switch strings.ToLower(c.options.DatabaseDriver) {
	case MySQL, MariaDB:
		c.options.DatabaseDriver = MySQL
	case SQLite3, "sqlite3", "test", "file", "":
		c.options.DatabaseDriver = SQLite3
	case Postgres:
		c.options.DatabaseDriver = Postgres
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

// DatabaseDriverName returns the formatted database driver name.
func (c *Config) DatabaseDriverName() string {
	switch c.DatabaseDriver() {
	case MySQL, MariaDB:
		return "MariaDB"
	case SQLite3, "sqlite3", "test", "file", "":
		return "SQLite"
	case Postgres:
		return "PostgreSQL"
	case "tidb":
		return "TiDB"
	default:
		return "unsupported database"
	}
}

// DatabaseVersion returns the database version string, if known.
func (c *Config) DatabaseVersion() string {
	return c.dbVersion
}

// IsDatabaseVersion checks if the database version is at least the specified version in semver format.
func (c *Config) IsDatabaseVersion(semverVersion string) bool {
	if semverVersion == "" {
		return true
	}

	return semver.Compare(c.DatabaseVersion(), semverVersion) >= 0
}

// DatabaseSsl checks if the database supports SSL connections for backup and restore.
func (c *Config) DatabaseSsl() bool {
	if c.dbVersion == "" {
		return false
	}

	switch c.DatabaseDriver() {
	case MySQL:
		// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
		return c.IsDatabaseVersion("v11.4")
	default:
		return false
	}
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
				"postgresql://%s:%s@%s:%d/%s?TimeZone=UTC&connect_timeout=%d&lock_timeout=50000&sslmode=disable",
				c.DatabaseUser(),
				c.DatabasePassword(),
				c.DatabaseHost(),
				c.DatabasePort(),
				c.DatabaseName(),
				c.DatabaseTimeout(),
			)
		case SQLite3:
			return filepath.Join(c.StoragePath(), "index.db?_busy_timeout=5000&_foreign_keys=on")
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
		return localhost
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

// Get the port based on the database driver Postgres vs MySQL/MariaDB
func (c *Config) _DefaultDatabasePort() int {
	if c.DatabaseDriver() == Postgres {
		return 5432
	}
	return 3306
}

// DatabasePort the database server port.
func (c *Config) DatabasePort() int {
	defaultPort := c._DefaultDatabasePort()

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

	// Try to read password from file if c.options.DatabasePassword is not set.
	if c.options.DatabasePassword != "" {
		return clean.Password(c.options.DatabasePassword)
	} else if fileName := FlagFilePath("DATABASE_PASSWORD"); fileName == "" {
		// No password set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
		log.Warnf("config: failed to read database password from %s (%s)", fileName, err)
		return ""
	} else {
		return clean.Password(string(b))
	}
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
		log.Debugf(fmt.Sprintf("Stack Trace: %s", debug.Stack()))
		log.Fatal("config: database not connected")
	}

	return c.db
}

// CloseDb closes the db connection (if any).
func (c *Config) CloseDb() error {
	if c.db != nil {
		sqldb, dberr := c.db.DB()
		if dberr != nil {
			sqldb.Close()
			c.db = nil
			entity.SetDbProvider(nil)
		} else {
			return dberr
		}
		if c.pool != nil {
			c.pool.Close()
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

// MigrateDb will initialize the database and migrate the schema if necessary.
func (c *Config) MigrateDb(runFailed bool, ids []string) {
	entity.Admin.UserName = c.AdminUser()

	// Automatically migrate database schema only once per release to reduce startup time.
	version := migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.Edition()))
	entity.InitDb(migrate.Opt(version.NeedsMigration(), runFailed, ids))
	if err := version.Migrated(c.Db()); err != nil {
		log.Warnf("config: %s (migrate)", err)
	}

	// Set the password for the initial Super Admin account, if specified.
	if c.AdminPassword() == "" {
		log.Warnf("config: %s account cannot be initialized due to missing or invalid password", clean.LogQuote(c.AdminUser()))
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword())
	}

	// Start recording warnings and errors after the required database table has been created.
	entity.LogWarningsAndErrors()
}

// InitTestDb drops all tables in the currently configured database and re-creates them.
func (c *Config) InitTestDb() {
	entity.ResetTestFixtures()

	if c.AdminPassword() == "" {
		// Do nothing.
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword())
	}

	// Start recording warnings and errors after the required table has have been created.
	entity.LogWarningsAndErrors()
}

// checkDb checks the database server version.
func (c *Config) checkDb(db *gorm.DB) error {
	switch c.DatabaseDriver() {
	case MySQL:
		type Res struct {
			Value string `gorm:"column:Value;"`
		}

		var res Res

		err := db.Raw("SELECT VERSION() AS Value").Scan(&res).Error

		if err != nil {
			err = db.Raw("SHOW VARIABLES LIKE 'innodb_version'").Scan(&res).Error
		}

		// Version query not supported.
		if err != nil {
			log.Tracef("config: failed to detect database version (%s)", err)
			return nil
		}

		c.dbVersion = clean.Version(res.Value)

		if c.dbVersion == "" {
			log.Warnf("config: unknown database server version")
		} else if !c.IsDatabaseVersion("v10.0.0") {
			return fmt.Errorf("config: MySQL %s is not supported, see https://docs.photoprism.app/getting-started/#databases", c.dbVersion)
		} else if !c.IsDatabaseVersion("v10.5.12") {
			return fmt.Errorf("config: MariaDB %s is not supported, see https://docs.photoprism.app/getting-started/#databases", c.dbVersion)
		}
	case Postgres:
		var versions []string
		err := db.Raw("SELECT VERSION() AS Value").Pluck("value", &versions).Error
		// Version query not supported.
		if err != nil {
			log.Tracef("config: failed to detect database version (%s)", err)
			return nil
		}

		c.dbVersion = clean.Version(versions[0])

		if c.dbVersion == "" {
			log.Warnf("config: unknown database server version")
		}
	case SQLite3:
		type Res struct {
			Value string `gorm:"column:Value;"`
		}

		var res Res

		err := db.Raw("SELECT sqlite_version() AS Value").Scan(&res).Error

		// Version query not supported.
		if err != nil {
			log.Warnf("config: failed to detect database version (%s)", err)
			return nil
		}

		c.dbVersion = clean.Version(res.Value)

		if c.dbVersion == "" {
			log.Warnf("config: unknown database server version")
		}
	}

	return nil
}

// Configure database logging.
func gormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.New(
			log, // This should be dummy.NewLogger(), to match GORM1.  Set to log before release...
			logger.Config{
				SlowThreshold:             time.Second,  // Slow SQL threshold
				LogLevel:                  logger.Error, // Log level  <-- This should be Silent to match GORM1, set to Error before release...
				IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      true,         // Don't include params in the SQL log
				Colorful:                  false,        // Disable color
			},
		),
		// Set UTC as the default for created and updated timestamps.
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
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
	var db *gorm.DB
	var err error
	if dbDriver == Postgres {
		postgresDB, pgxPool := entity.OpenPostgreSQL(dbDsn)
		c.pool = pgxPool
		db, err = gorm.Open(postgres.New(postgres.Config{Conn: postgresDB}), gormConfig())
	} else {
		c.pool = nil
		db, err = gorm.Open(drivers[dbDriver](dbDsn), gormConfig())
	}
	if err != nil || db == nil {
		log.Infof("config: waiting for the database to become available")

		for i := 1; i <= 12; i++ {
			if dbDriver == Postgres {
				postgresDB, pgxPool := entity.OpenPostgreSQL(dbDsn)
				c.pool = pgxPool
				db, err = gorm.Open(postgres.New(postgres.Config{Conn: postgresDB}), gormConfig())
			} else {
				c.pool = nil
				db, err = gorm.Open(drivers[dbDriver](dbDsn), gormConfig())
			}

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
	//db.LogMode(false)
	//db.SetLogger(log)

	// Set database connection parameters.
	if dbDriver != Postgres {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxOpenConns(c.DatabaseConns())
		sqlDB.SetMaxIdleConns(c.DatabaseConnsIdle())
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// Check database server version.
	if err = c.checkDb(db); err != nil {
		if c.Unsafe() {
			log.Error(err)
		} else {
			return err
		}
	}

	if dbVersion := c.DatabaseVersion(); dbVersion != "" {
		log.Debugf("database: opened connection to %s %s", c.DatabaseDriverName(), dbVersion)
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
