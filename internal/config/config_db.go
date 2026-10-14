package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/mod/semver"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

// postgresExperimentalOnce reports the experimental status of PostgreSQL support once.
var postgresExperimentalOnce sync.Once

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
	case dsn.DriverPostgres:
		postgresExperimentalOnce.Do(func() {
			log.Infof("config: support for PostgreSQL is experimental")
		})
		c.options.DatabaseDriver = dsn.DriverPostgres
	case dsn.DriverSQLite3, dsn.DriverNone, dsn.DriverAuto:
		c.options.DatabaseDriver = dsn.DriverSQLite3
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
	case dsn.DriverPostgres:
		return "PostgreSQL"
	case dsn.DriverSQLite3:
		return "SQLite"
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
			d := dsn.DSN{
				User:     c.DatabaseUser(),
				Password: c.DatabasePassword(),
				Net:      "tcp",
				Server:   c.DatabaseServer(),
				Name:     c.DatabaseName(),
				Params:   fmt.Sprintf("%s&timeout=%ds", dsn.Params[dsn.DriverMySQL], c.DatabaseTimeout()),
			}

			// Connect via Unix Domain Socket?
			if strings.HasPrefix(d.Server, "/") {
				log.Debugf("mariadb: connecting via Unix domain socket")
				d.Net = "unix"
			}

			return d.MySQL()
		case dsn.DriverPostgres:
			d := dsn.DSN{
				Driver:   dsn.DriverPostgres,
				User:     c.DatabaseUser(),
				Password: c.DatabasePassword(),
				Name:     c.DatabaseName(),
				Params:   fmt.Sprintf("%s connect_timeout=%d", dsn.Params[dsn.DriverPostgres], c.DatabaseTimeout()),
			}

			// Connect via Unix domain socket, or to the host and port of the server.
			if server := c.DatabaseServer(); strings.HasPrefix(server, "/") {
				d.Server = server

				// A socket path may end with the port, e.g. "/var/run/postgresql:5433".
				if i := strings.LastIndex(server, ":"); i > 0 {
					if port, err := strconv.Atoi(server[i+1:]); err == nil && port > 0 && port <= 65535 {
						d.Server = server[:i]
						d.Params += fmt.Sprintf(" port=%d", port)
					}
				}
			} else {
				hostPort := dsn.DSN{Driver: dsn.DriverPostgres, Server: server}
				d.Server = net.JoinHostPort(hostPort.Host(), strconv.Itoa(hostPort.Port()))
			}

			return d.PostgreSQL()
		case dsn.DriverSQLite3:
			return filepath.Join(c.StoragePath(), fmt.Sprintf("index.db?%s", dsn.Params[dsn.DriverSQLite3]))
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

	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return ""
	}

	d := c.databaseServerDSN()
	return d.Host()
}

// DatabasePort the database server port.
func (c *Config) DatabasePort() int {
	c.ParseDatabaseDSN()

	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		return 0
	}

	d := c.databaseServerDSN()
	return d.Port()
}

// databaseServerDSN returns the DSN that holds the database server address: the configured DSN, or,
// without one, the configured server, so a password is never parsed as part of the address.
func (c *Config) databaseServerDSN() dsn.DSN {
	if c.HasDatabaseDSN() {
		return dsn.Parse(c.DatabaseDSN())
	}

	return dsn.DSN{Driver: c.DatabaseDriver(), Server: c.DatabaseServer()}
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
	} else if c.dbNoPassFile {
		return ""
	} else if fileName := FlagFilePath("DATABASE_PASSWORD"); fileName == "" {
		// No password set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 { //nolint:gosec // path derived from environment variable for DB password
		event.SystemWarn([]string{"config", "database password", "read %s", "%s"}, clean.Log(fileName), clean.ErrorFull(err))
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
		log.Fatal("config: database not connected")
	}

	return c.db
}

// DbIfConnected returns the database connection, or nil if it is not connected, for callers that can
// continue without it rather than exit like Db.
func (c *Config) DbIfConnected() *gorm.DB {
	if c == nil {
		return nil
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
		if err := entity.CloseDb(c.db); err == nil {
			c.db = nil
			entity.SetDbProvider(nil)
		} else {
			return err
		}
	}

	return nil
}

// IsDbOpen reports whether the database is connected and responding.
func (c *Config) IsDbOpen() bool {
	if c.db == nil {
		log.Debug("config: database not connected")
		return false
	}

	if sqlDb, err := c.db.DB(); err != nil {
		log.Debugf("config: database not available (%s)", err)
		return false
	} else if err = sqlDb.Ping(); err != nil {
		log.Debugf("config: database not available (%s)", err)
		return false
	}

	return true
}

// RegisterDb opens a database connection if needed and sets the connection provider.
func (c *Config) RegisterDb() {
	if err := c.connectDb(); err != nil {
		// Report via the system log, not the database-persisted logger, so a
		// connection failure cannot trigger a follow-up error writing to the DB.
		event.SystemError([]string{"config", "database", "register", "%s"}, clean.ErrorFull(err))
		return
	}

	entity.SetDbProvider(c)
}

// versionEdition returns the edition stored with the schema version. Portal builds get rows of their own,
// as they register tables that other editions of the same release sharing a database do not.
func (c *Config) versionEdition() string {
	edition := c.Edition()

	if c.Portal() && edition != Portal {
		return Portal + "-" + edition
	}

	return edition
}

// InitDb initializes the database without running previously failed migrations.
func (c *Config) InitDb() {
	c.RegisterDb()

	// MigrateDb logs the error and records it on a known version.
	_ = c.MigrateDb(false, nil)
}

// MigrateDb will initialize the database and migrate the schema if necessary, and returns an error if
// the schema could not be migrated.
func (c *Config) MigrateDb(runFailed bool, ids []string) (err error) {
	// A schema that migrations would alter is refused before anything writes, also after a restore.
	if err = entity.CheckPostgresSchema(c.Db()); err != nil {
		log.Errorf("config: %s", clean.Error(err))
		return err
	}

	entity.Admin.UserName = c.AdminUser()

	// Automatically migrate database schema only once per release to reduce startup time.
	version := migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.versionEdition()))
	opt := migrate.Opt(version.NeedsMigration(), runFailed, ids)

	// Record the version as migrated only once its schema was, so the next start migrates it again.
	if err = entity.InitDb(opt); err != nil {
		event.SystemError([]string{"config", "%s (migrate)"}, clean.ErrorFull(err))

		if !version.Unknown() {
			version.MigratedAt = nil
			version.Error = clean.ErrorBytes(err, 255)

			if saveErr := version.Save(c.Db()); saveErr != nil {
				event.SystemWarn([]string{"config", "%s (save version)"}, clean.ErrorFull(saveErr))
			}
		}
	} else if opt.AutoMigrate {
		// Running selected migrations only does not migrate the schema, so it records nothing.
		if migratedErr := version.Migrated(c.Db()); migratedErr != nil {
			event.SystemWarn([]string{"config", "%s (save version)"}, clean.ErrorFull(migratedErr))
		}
	}

	// Set the password for the initial Super Admin account, if specified.
	if c.AdminPassword() == "" {
		log.Warnf("config: %s account cannot be initialized due to missing or invalid password", clean.LogQuote(c.AdminUser()))
	} else {
		entity.Admin.InitAccount(c.AdminInitName(), c.AdminPassword(), c.AdminScope())
	}

	// Start recording warnings and errors after the required database table has been created.
	entity.LogWarningsAndErrors()

	return err
}

// MigrateDbSchema migrates the database schema like MigrateDb, without creating default records or using the
// database of the entity package, e.g. for the target of a transfer.
func (c *Config) MigrateDbSchema() error {
	db := c.Db()

	if err := entity.CheckPostgresSchema(db); err != nil {
		return err
	}

	// The migration and version helpers create their tables only once per process, so they are created here.
	for _, table := range []any{&migrate.Migration{}, &migrate.Version{}} {
		if db.Migrator().HasTable(table) {
			continue
		} else if err := db.AutoMigrate(table); err != nil {
			return err
		}
	}

	version := migrate.FirstOrCreateVersion(db, migrate.NewVersion(c.Version(), c.versionEdition()))

	if err := entity.Entities.Migrate(db, migrate.Opt(true, false, nil)); err != nil {
		return err
	} else if err = entity.Entities.WaitForMigration(db); err != nil {
		return err
	}

	return version.Migrated(db)
}

// InitTestDb drops all tables in the currently configured database and re-creates them.
func (c *Config) InitTestDb() {
	// Make sure that the migrations and versions tables are already there, as once prevents these from being handled correctly in tests.
	if !c.db.Migrator().HasTable(&migrate.Migration{}) {
		if err := c.db.AutoMigrate(&migrate.Migration{}); err != nil {
			log.Errorf("config: %s (create migrations table)", clean.Error(err))
		}
	}
	if !c.db.Migrator().HasTable(&migrate.Version{}) {
		if err := c.db.AutoMigrate(&migrate.Version{}); err != nil {
			log.Errorf("config: %s (create versions table)", clean.Error(err))
		}
	}
	entity.ResetTestFixtures()

	if c.AdminPassword() == "" {
		// Do nothing.
	} else {
		entity.Admin.InitAccount(c.AdminInitName(), c.AdminPassword(), c.AdminScope())
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
	case dsn.DriverPostgres:
		var versionNum int

		// Version query not supported.
		if err := db.Raw("SELECT current_setting('server_version_num')::integer").Row().Scan(&versionNum); err != nil {
			log.Warnf("config: failed to detect database version (%s)", err)
			return nil
		}

		c.dbVersion = postgresVersion(versionNum)

		switch {
		case c.dbVersion == "":
			log.Warnf("config: unknown database server version")
		case !c.IsDatabaseVersion(PostgresMinVersion):
			return fmt.Errorf("PostgreSQL %s is not supported, see https://docs.photoprism.app/getting-started/#databases", c.dbVersion)
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

// PostgresMinVersion is the oldest PostgreSQL release that is supported.
const PostgresMinVersion = "v16"

// postgresVersion returns the semantic version of a PostgreSQL server_version_num, e.g. v18.6 for 180006
// and v9.6 for 90624, or an empty string if the number is not a version.
func postgresVersion(num int) string {
	switch {
	case num >= 100000:
		return fmt.Sprintf("v%d.%d", num/10000, num%10000)
	case num >= 10000:
		return fmt.Sprintf("v%d.%d", num/10000, num/100%100)
	default:
		return ""
	}
}

// DatabaseOptions holds the connection settings of a database other than the configured one.
type DatabaseOptions struct {
	Driver   string
	DSN      string
	Server   string
	Name     string
	User     string
	Password string //nolint:gosec // G117: connection settings intentionally include the password.
}

// WithDatabase returns a copy of the config that connects to the specified database instead, e.g. as the
// target of a transfer. Settings that are not specified get their defaults rather than the configured values,
// and the configured password file does not apply.
func (c *Config) WithDatabase(db DatabaseOptions) *Config {
	opt := *c.Options()
	opt.DatabaseDriver = db.Driver
	opt.DatabaseDSN = db.DSN
	opt.DatabaseServer = db.Server
	opt.DatabaseName = db.Name
	opt.DatabaseUser = db.User
	opt.DatabasePassword = db.Password
	opt.Deprecated.DatabaseDsn = ""

	return &Config{cliCtx: c.cliCtx, options: &opt, settings: c.settings, env: c.env, dbNoPassFile: true}
}

// ConnectDb opens the database connection like RegisterDb, but returns an error if that fails and does not
// make it the connection of the entity package.
func (c *Config) ConnectDb() error {
	return c.connectDb()
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
	dbDsn := c.DatabaseDSN()

	if dbDriver == "" {
		return errors.New("driver not specified")
	}

	if dbDsn == "" {
		return errors.New("DSN not specified")
	}

	// Fail without retrying if the driver is not supported.
	if _, err := entity.GormDialector(dbDriver, dbDsn, nil); err != nil {
		return err
	}

	// Open database connection.
	db, err := entity.OpenDb(dbDriver, dbDsn)
	if err != nil || db == nil {
		log.Infof("config: waiting for the database to become available")

		for i := 1; i <= 12; i++ {
			db, err = entity.OpenDb(dbDriver, dbDsn)

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
	if sqlDb, sqlErr := db.DB(); sqlErr != nil {
		_ = entity.CloseDb(db)
		return sqlErr
	} else {
		sqlDb.SetMaxOpenConns(c.DatabaseConns())
		sqlDb.SetMaxIdleConns(c.DatabaseConnsIdle())
		sqlDb.SetConnMaxLifetime(time.Hour)
	}

	// Check database server version.
	if err = c.checkDb(db); err != nil {
		if c.Unsafe() {
			// Report via the system log so a database problem is not written to
			// the database-persisted error log.
			event.SystemError([]string{"config", "database", "check", "%s"}, clean.ErrorFull(err))
		} else {
			_ = entity.CloseDb(db)
			return err
		}
	}

	// A schema that migrations would alter is refused regardless of the checks above.
	if err = entity.CheckPostgresSchema(db); err != nil {
		_ = entity.CloseDb(db)
		return err
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
	contents, err := os.ReadFile(filename) //nolint:gosec // import path is provided by trusted caller

	if err != nil {
		log.Error(err)
		return
	}

	statements := strings.Split(string(contents), ";\n")
	q := c.Db().Unscoped().Session(&gorm.Session{})

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
