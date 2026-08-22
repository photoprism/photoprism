package config

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Auto requests automatic detection of an implementation-defined default
// (e.g. the database driver). The canonical SQL driver identifiers live in
// pkg/dsn (dsn.DriverMySQL, dsn.DriverSQLite3, …).
const Auto = "auto"

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	c.normalizeDatabaseDSN()

	switch dsn.ParseDriver(c.options.DatabaseDriver) {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		c.options.DatabaseDriver = dsn.DriverMySQL
	case dsn.DriverSQLite3, dsn.DriverNone, dsn.DriverAuto:
		c.options.DatabaseDriver = dsn.DriverSQLite3
	case dsn.DriverPostgreSQL, dsn.DriverPostgres:
		c.options.DatabaseDriver = dsn.DriverPostgreSQL
	case dsn.DriverTiDB:
		log.Warnf("config: database driver 'tidb' is deprecated, using sqlite")
		c.options.DatabaseDriver = dsn.DriverSQLite3
		c.options.DatabaseDSN = ""
	default:
		log.Warnf("config: unsupported database driver %s, using sqlite", c.options.DatabaseDriver)
		c.options.DatabaseDriver = dsn.DriverSQLite3
		c.options.DatabaseDSN = ""
	}

	return c.options.DatabaseDriver
}

// DatabaseDriverName returns the formatted database driver name. Input is
// always canonical after DatabaseDriver(); the default arm is defensive.
func (c *Config) DatabaseDriverName() string {
	switch c.DatabaseDriver() {
	case dsn.DriverMySQL:
		return "MariaDB"
	case dsn.DriverSQLite3:
		return "SQLite"
	case dsn.DriverPostgreSQL:
		return "PostgreSQL"
	case dsn.DriverAuto:
		return "Auto"
	default:
		return "Unsupported"
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
	case dsn.DriverMySQL:
		// see https://mariadb.org/mission-impossible-zero-configuration-ssl/
		return c.IsDatabaseVersion("v11.4")
	default:
		return false
	}
}

// normalizeDatabaseDSN maps the deprecated DatabaseDsn database configuration
// value to its current counterpart, DatabaseDSN, before consumption.
func (c *Config) normalizeDatabaseDSN() {
	if c.options.DatabaseDSN == "" && c.options.Deprecated.DatabaseDsn != "" {
		c.options.DatabaseDSN = c.options.Deprecated.DatabaseDsn
		c.options.Deprecated.DatabaseDsn = ""
		event.SystemWarn([]string{"config", "options", "DatabaseDsn has been deprecated in favor of DatabaseDSN"})
	}
}

// DatabaseDSN returns the database data source name (DSN).
func (c *Config) DatabaseDSN() string {
	// Generate matching database DSN based on the configured database driver.
	if c.NoDatabaseDSN() {
		switch c.DatabaseDriver() {
		case dsn.DriverMySQL:
			databaseNet := "tcp"
			// Connect via Unix Domain Socket?
			if strings.HasPrefix(c.DatabaseServer(), "/") {
				log.Debugf("mariadb: connecting via Unix domain socket")
				databaseNet = "unix"
			}
			return (&dsn.DSN{
				Driver:   dsn.DriverMySQL,
				User:     c.DatabaseUser(),
				Password: c.DatabasePassword(),
				Server:   c.DatabaseServer(),
				Net:      databaseNet,
				Name:     c.DatabaseName(),
				Params:   fmt.Sprintf("%s&timeout=%ds", dsn.Params[dsn.DriverMySQL], c.DatabaseTimeout()),
			}).ToString()
		case dsn.DriverPostgres, dsn.DriverPostgreSQL:
			return (&dsn.DSN{
				Driver:   dsn.DriverPostgreSQL,
				User:     c.DatabaseUser(),
				Password: c.DatabasePassword(),
				Server:   c.DatabaseServer(),
				Name:     c.DatabaseName(),
				Params:   fmt.Sprintf("connect_timeout=%d&%s", c.DatabaseTimeout(), dsn.Params[dsn.DriverPostgreSQL]),
			}).ToString()
		case dsn.DriverSQLite3:
			return (&dsn.DSN{
				Driver: dsn.DriverSQLite3,
				Server: c.StoragePath(),
				Name:   "index.db",
				Params: fmt.Sprintf("%s", dsn.Params[dsn.DriverSQLite3]),
			}).ToString()
		default:
			log.Errorf("config: empty database dsn")
			return ""
		}
	}

	// If missing, add the required parameters to the configured MySQL/MariaDB DSN.
	if c.DatabaseDriver() == dsn.DriverMySQL && !strings.Contains(c.options.DatabaseDSN, "?") {
		c.options.DatabaseDSN = fmt.Sprintf(
			"%s?%s&timeout=%ds",
			c.options.DatabaseDSN,
			dsn.Params[dsn.DriverMySQL],
			c.DatabaseTimeout())
	}

	return c.options.DatabaseDSN
}

// NoDatabaseDSN checks if no manual database data source name (DSN) configuration is set.
func (c *Config) NoDatabaseDSN() bool {
	c.normalizeDatabaseDSN()

	return c.options.DatabaseDSN == ""
}

// HasDatabaseDSN checks if a manual database data source name (DSN) configuration is set.
func (c *Config) HasDatabaseDSN() bool {
	return !c.NoDatabaseDSN()
}

// ReportDatabaseDSN checks if the database data source name (DSN) should be reported
// instead of database name, server, user, and password.
func (c *Config) ReportDatabaseDSN() bool {
	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return true
	}

	return c.HasDatabaseDSN()
}

// ParseDatabaseDSN parses the database dsn and extracts user, password, database server, and name.
func (c *Config) ParseDatabaseDSN() {
	if c.NoDatabaseDSN() {
		return
	} else if c.options.DatabaseServer != "" && c.DatabaseDriver() == dsn.DriverSQLite3 {
		return
	}

	d := dsn.Parse(c.options.DatabaseDSN)

	c.options.DatabaseName = d.Name
	c.options.DatabaseServer = d.Server
	c.options.DatabaseUser = d.User
	c.options.DatabasePassword = d.Password
}

// DatabaseFile returns the filename part of a sqlite database DSN.
func (c *Config) DatabaseFile() string {
	fileName, _, _ := strings.Cut(strings.TrimPrefix(c.DatabaseDSN(), "file:"), "?")
	return fileName
}

// DatabaseServer the database server.
func (c *Config) DatabaseServer() string {
	c.ParseDatabaseDSN()

	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return ""
	} else if c.options.DatabaseServer == "" {
		return localhost
	}

	return c.options.DatabaseServer
}

// DatabaseHost the database server host.
func (c *Config) DatabaseHost() string {
	c.ParseDatabaseDSN()

	if c.DatabaseDriver() == dsn.DriverSQLite3 || c.NoDatabaseDSN() {
		d := dsn.DSN{Driver: c.DatabaseDriver(), Server: c.DatabaseServer(), DSN: ""}
		return d.Host()
	}

	d := dsn.Parse(c.DatabaseDSN())
	return d.Host()
}

// DatabasePort the database server port.
func (c *Config) DatabasePort() int {
	c.ParseDatabaseDSN()

	if c.DatabaseDriver() == dsn.DriverSQLite3 || c.NoDatabaseDSN() {
		d := dsn.DSN{Driver: c.DatabaseDriver(), Server: c.DatabaseServer(), DSN: ""}
		return d.Port()
	}

	d := dsn.Parse(c.DatabaseDSN())
	return d.Port()
}

// DatabasePortString the database server port as string.
func (c *Config) DatabasePortString() string {
	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return ""
	}

	return strconv.Itoa(c.DatabasePort())
}

// DatabaseName the database schema name.
func (c *Config) DatabaseName() string {
	c.ParseDatabaseDSN()

	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return c.DatabaseDSN()
	} else if c.options.DatabaseName == "" {
		return "photoprism"
	}

	return c.options.DatabaseName
}

// DatabaseUser returns the database user name.
func (c *Config) DatabaseUser() string {
	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return ""
	}

	c.ParseDatabaseDSN()

	if c.options.DatabaseUser == "" {
		return "photoprism"
	}

	return c.options.DatabaseUser
}

// DatabasePassword returns the database user password.
func (c *Config) DatabasePassword() string {
	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return ""
	}

	c.ParseDatabaseDSN()

	// Try to read password from file if c.options.DatabasePassword is not set.
	if c.options.DatabasePassword != "" {
		return clean.Password(c.options.DatabasePassword)
	} else if fileName := FlagFilePath("DATABASE_PASSWORD"); fileName == "" {
		// No password set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 { //nolint:gosec // path derived from environment variable for DB password
		event.SystemWarn([]string{"config", "database password", "read %s", "%s"}, clean.Log(fileName), clean.Error(err))
		return ""
	} else {
		return clean.Password(string(b))
	}
}

// DatabaseProvisionPrefix returns the sanitized prefix for provisioned database names and users.
func (c *Config) DatabaseProvisionPrefix() string {
	prefix := strings.TrimSpace(c.options.DatabaseProvisionPrefix)

	if prefix == "" {
		return cluster.DefaultDatabaseProvisionPrefix
	}

	prefix = strings.ToLower(prefix)

	cleaned := make([]rune, 0, len(prefix))
	prevUnderscore := false

	for _, r := range prefix {
		switch {
		case r >= 'a' && r <= 'z':
			cleaned = append(cleaned, r)
			prevUnderscore = false
		case r >= '0' && r <= '9':
			if len(cleaned) == 0 {
				continue
			}
			cleaned = append(cleaned, r)
			prevUnderscore = false
		case r == '_' || r == '-' || r == ' ':
			if len(cleaned) == 0 || prevUnderscore {
				continue
			}
			cleaned = append(cleaned, '_')
			prevUnderscore = true
		default:
			continue
		}

		if len(cleaned) >= cluster.DatabaseProvisionPrefixMaxLen {
			break
		}
	}

	if len(cleaned) == 0 {
		return cluster.DefaultDatabaseProvisionPrefix
	}

	result := string(cleaned)
	c.options.DatabaseProvisionPrefix = result

	return result
}

// ShouldAutoRotateDatabase decides whether callers should request DB rotation automatically.
// It is used by both the CLI and node bootstrap to avoid unnecessary provisioning calls.
func (c *Config) ShouldAutoRotateDatabase() bool {
	if c.Portal() || c.DatabaseDriver() != dsn.DriverMySQL {
		return false
	}

	if c.DatabaseName() == "" || c.DatabaseUser() == "" || c.DatabasePassword() == "" {
		return true
	}

	return false
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

// AsyncJobDrainTimeout bounds how long CloseDb waits for background jobs before
// tearing down the connection, so a wedged job cannot hang shutdown forever.
const AsyncJobDrainTimeout = 30 * time.Second

// CloseDb closes the db connection (if any). It first drains async work registered
// with entity.AsyncJobAdd (UpdateCountsAsync, UpdateCoversAsync, …) so those goroutines
// do not race the provider being nilled, bounded by AsyncJobDrainTimeout so a stuck job
// degrades to a logged warning instead of an indefinite hang.
func (c *Config) CloseDb() error {
	// Reported on the console-only system log, as the database backing the error log is going away.
	if !entity.WaitForAsyncJobsTimeout(AsyncJobDrainTimeout) {
		event.SystemWarn([]string{"config", "database", "close", "timeout waiting for background jobs"})
	}

	if c.db != nil {
		sqldb, dberr := c.db.DB()
		if dberr == nil {
			log.Debug("config: closing database")
			if err := sqldb.Close(); err != nil {
				return err
			}
			entity.SetDbProvider(nil)
			c.db = nil
		} else {
			return dberr
		}
		if c.pool != nil {
			log.Debug("config: closing postgres pool")
			c.pool.Close()
			c.pool = nil
		}
	}

	return nil
}

// IsDbOpen determines if the database is available to use
func (c *Config) IsDbOpen() bool {
	if c.db == nil {
		log.Debug("config: database not connected")
		return false
	} else {
		if sqlDB, err := c.db.DB(); err != nil {
			log.Debugf("config: database not available (%s)", err)
			return false
		} else {
			if sqlErr := sqlDB.Ping(); sqlErr != nil {
				log.Debugf("config: database not available (%s)", sqlErr)
				return false
			} else {
				return true
			}
		}
	}
}

// SetDbOptions sets the database collation to unicode if supported.
func (c *Config) SetDbOptions() {
	switch c.DatabaseDriver() {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		c.Db().Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC")
	case dsn.DriverPostgres, dsn.DriverPostgreSQL:
		// Ignore for now.
	case dsn.DriverSQLite3:
		// Not required as Unicode is default.
	}
}

// RegisterDb opens a database connection if needed,
// sets the database options and connection provider.
func (c *Config) RegisterDb() {
	if err := c.connectDb(); err != nil {
		// Report via the system log, not the database-persisted logger, so a
		// connection failure cannot trigger a follow-up error writing to the DB.
		event.SystemError([]string{"config", "database", "register", "%s"}, clean.Error(err))
		return
	}

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
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword(), c.AdminScope())
	}

	// Start recording warnings and errors after the required database table has been created.
	entity.LogWarningsAndErrors()
}

// InitTestDb drops all tables in the currently configured database and re-creates them.
func (c *Config) InitTestDb() {
	// Make sure that the migrations and versions tables are already there, as once prevents these from being handled correctly in tests.
	if (!c.db.Migrator().HasTable(&migrate.Migration{})) {
		c.db.Migrator().AutoMigrate(&migrate.Migration{})
	}
	if (!c.db.Migrator().HasTable(&migrate.Version{})) {
		c.db.Migrator().AutoMigrate(&migrate.Version{})
	}
	entity.ResetTestFixtures()

	if c.AdminPassword() == "" {
		// Do nothing.
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword(), c.AdminScope())
	}

	// Start recording warnings and errors after the required table has have been created.
	entity.LogWarningsAndErrors()
}

// checkDb checks the database server version.
func (c *Config) checkDb(db *gorm.DB) error {
	if txt.Bool(os.Getenv(EnvVar("DATABASE_SKIP_VERSION_CHECK"))) {
		log.Debugf("config: skipping database version check")
		return nil
	}

	if db == nil {
		return fmt.Errorf("config: missing database connection")
	}

	switch c.DatabaseDriver() {
	case dsn.DriverMySQL:
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

		switch {
		case c.dbVersion == "":
			log.Warnf("config: unknown database server version")
		case !c.IsDatabaseVersion("v10.0.0"):
			return fmt.Errorf("MySQL %s is not supported, see https://docs.photoprism.app/getting-started/#databases", c.dbVersion)
		case !c.IsDatabaseVersion("v10.5.12"):
			return fmt.Errorf("MariaDB %s is not supported, see https://docs.photoprism.app/getting-started/#databases", c.dbVersion)
		}
	case dsn.DriverPostgres, dsn.DriverPostgreSQL:
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
	case dsn.DriverSQLite3:
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

	// Database connection already exists.
	if c.db != nil {
		return nil
	}

	// Get database driver and data source name.
	dbDriver := c.DatabaseDriver()
	dbDSN := c.DatabaseDSN()

	if dbDriver == "" {
		return errors.New("driver not specified")
	}

	if dbDSN == "" {
		return errors.New("DSN not specified")
	}

	if c.IsDbOpen() {
		log.Info("config: database is already open")
	} else {

		// Open database connection.
		var db *gorm.DB
		var err error
		if dbDriver == dsn.DriverPostgres || dbDriver == dsn.DriverPostgreSQL {
			postgresDB, pgxPool := entity.OpenPostgreSQL(dbDSN)
			c.pool = pgxPool
			db, err = gorm.Open(postgres.New(postgres.Config{Conn: postgresDB}), gormConfig())
		} else {
			c.pool = nil
			db, err = gorm.Open(dsn.GormDrivers[dbDriver](dbDSN), gormConfig())
		}
		if err != nil || db == nil {
			log.Infof("config: waiting for the database to become available")

			for i := 1; i <= 12; i++ {
				if dbDriver == dsn.DriverPostgres || dbDriver == dsn.DriverPostgreSQL {
					postgresDB, pgxPool := entity.OpenPostgreSQL(dbDSN)
					c.pool = pgxPool
					db, err = gorm.Open(postgres.New(postgres.Config{Conn: postgresDB}), gormConfig())
				} else {
					c.pool = nil
					db, err = gorm.Open(dsn.GormDrivers[dbDriver](dbDSN), gormConfig())
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

		// Set database connection parameters.
		if dbDriver != dsn.DriverPostgres && dbDriver != dsn.DriverPostgreSQL {
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
				// Report via the system log so a database problem is not written to
				// the database-persisted error log.
				event.SystemError([]string{"config", "database", "check", "%s"}, clean.Error(err))
			} else {
				return err
			}
		}

		if dbVersion := c.DatabaseVersion(); dbVersion != "" {
			log.Debugf("database: opened connection to %s %s", c.DatabaseDriverName(), dbVersion)
		}

		// Ok.
		c.db = db
	}

	return nil
}

// ImportSQL imports a file to the currently configured database.
// All lines, including comments, must be terminated with a ;\n
func (c *Config) ImportSQL(filename string) {
	contents, err := os.ReadFile(filename) //nolint:gosec // import path is provided by trusted caller

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

		if err := q.Exec(stmt).Error; err != nil {
			log.Error(err)
			return
		}
	}
}
