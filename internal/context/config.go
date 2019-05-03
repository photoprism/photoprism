package context

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/urfave/cli"
)

const (
	DbTiDB  = "internal"
	DbMySQL = "mysql"
)

// Config provides a struct in which application configuration is stored.
// Application code must use functions to get config values, for two reasons:
//
// 1. Some values are computed and we don't want to leak implementation details (aims at reducing refactoring overhead).
//
// 2. Paths might actually be dynamic later (if we build a multi-user version).
//
// See https://github.com/photoprism/photoprism/issues/50#issuecomment-433856358
type Config struct {
	Name               string
	Version            string
	Copyright          string
	Debug              bool
	LogLevel           string
	ConfigFile         string
	AssetsPath         string
	CachePath          string
	OriginalsPath      string
	ImportPath         string
	ExportPath         string
	SqlServerHost      string
	SqlServerPort      uint
	SqlServerPath      string
	SqlServerPassword  string
	HttpServerHost     string
	HttpServerPort     int
	HttpServerMode     string
	HttpServerPassword string
	DarktableCli       string
	DatabaseDriver     string
	DatabaseDsn        string
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (c *Config) SetValuesFromFile(fileName string) error {
	yamlConfig, err := yaml.ReadFile(fileName)

	if err != nil {
		return err
	}

	c.ConfigFile = fileName
	if debug, err := yamlConfig.GetBool("debug"); err == nil {
		c.Debug = debug
	}

	if logLevel, err := yamlConfig.Get("log-level"); err == nil {
		c.LogLevel = logLevel
	}

	if sqlServerHost, err := yamlConfig.Get("sql-host"); err == nil {
		c.SqlServerHost = sqlServerHost
	}

	if sqlServerPort, err := yamlConfig.GetInt("sql-port"); err == nil {
		c.SqlServerPort = uint(sqlServerPort)
	}

	if sqlServerPassword, err := yamlConfig.Get("sql-password"); err == nil {
		c.SqlServerPassword = sqlServerPassword
	}

	if sqlServerPath, err := yamlConfig.Get("sql-path"); err == nil {
		c.SqlServerPath = sqlServerPath
	}

	if httpServerHost, err := yamlConfig.Get("http-host"); err == nil {
		c.HttpServerHost = httpServerHost
	}

	if httpServerPort, err := yamlConfig.GetInt("http-port"); err == nil {
		c.HttpServerPort = int(httpServerPort)
	}

	if httpServerMode, err := yamlConfig.Get("http-mode"); err == nil {
		c.HttpServerMode = httpServerMode
	}

	if httpServerPassword, err := yamlConfig.Get("http-password"); err == nil {
		c.HttpServerPassword = httpServerPassword
	}

	if assetsPath, err := yamlConfig.Get("assets-path"); err == nil {
		c.AssetsPath = fsutil.ExpandedFilename(assetsPath)
	}

	if cachePath, err := yamlConfig.Get("cache-path"); err == nil {
		c.CachePath = fsutil.ExpandedFilename(cachePath)
	}

	if originalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		c.OriginalsPath = fsutil.ExpandedFilename(originalsPath)
	}

	if importPath, err := yamlConfig.Get("import-path"); err == nil {
		c.ImportPath = fsutil.ExpandedFilename(importPath)
	}

	if exportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.ExportPath = fsutil.ExpandedFilename(exportPath)
	}

	if darktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.DarktableCli = fsutil.ExpandedFilename(darktableCli)
	}

	if databaseDriver, err := yamlConfig.Get("database-driver"); err == nil {
		c.DatabaseDriver = databaseDriver
	}

	if databaseDsn, err := yamlConfig.Get("database-dsn"); err == nil {
		c.DatabaseDsn = databaseDsn
	}

	return nil
}

// SetValuesFromCliContext uses values from the CLI to setup configuration overrides
// for the entity.
func (c *Config) SetValuesFromCliContext(ctx *cli.Context) error {
	if ctx.GlobalBool("debug") {
		c.Debug = ctx.GlobalBool("debug")
	}

	if ctx.GlobalIsSet("log-level") || c.LogLevel == "" {
		c.LogLevel = ctx.GlobalString("log-level")
	}

	if ctx.GlobalIsSet("assets-path") || c.AssetsPath == "" {
		c.AssetsPath = fsutil.ExpandedFilename(ctx.GlobalString("assets-path"))
	}

	if ctx.GlobalIsSet("cache-path") || c.CachePath == "" {
		c.CachePath = fsutil.ExpandedFilename(ctx.GlobalString("cache-path"))
	}

	if ctx.GlobalIsSet("originals-path") || c.OriginalsPath == "" {
		c.OriginalsPath = fsutil.ExpandedFilename(ctx.GlobalString("originals-path"))
	}

	if ctx.GlobalIsSet("import-path") || c.ImportPath == "" {
		c.ImportPath = fsutil.ExpandedFilename(ctx.GlobalString("import-path"))
	}

	if ctx.GlobalIsSet("export-path") || c.ExportPath == "" {
		c.ExportPath = fsutil.ExpandedFilename(ctx.GlobalString("export-path"))
	}

	if ctx.GlobalIsSet("darktable-cli") || c.DarktableCli == "" {
		c.DarktableCli = fsutil.ExpandedFilename(ctx.GlobalString("darktable-cli"))
	}

	if ctx.GlobalIsSet("database-driver") || c.DatabaseDriver == "" {
		c.DatabaseDriver = ctx.GlobalString("database-driver")
	}

	if ctx.GlobalIsSet("database-dsn") || c.DatabaseDsn == "" {
		c.DatabaseDsn = ctx.GlobalString("database-dsn")
	}

	if ctx.GlobalIsSet("sql-host") || c.SqlServerHost == "" {
		c.SqlServerHost = ctx.GlobalString("sql-host")
	}

	if ctx.GlobalIsSet("sql-port") || c.SqlServerPort == 0 {
		c.SqlServerPort = ctx.GlobalUint("sql-port")
	}

	if ctx.GlobalIsSet("sql-password") || c.SqlServerPassword == "" {
		c.SqlServerPassword = ctx.GlobalString("sql-password")
	}

	if ctx.GlobalIsSet("sql-path") || c.SqlServerPath == "" {
		c.SqlServerPath = ctx.GlobalString("sql-path")
	}

	if ctx.GlobalIsSet("http-host") || c.HttpServerHost == "" {
		c.HttpServerHost = ctx.GlobalString("http-host")
	}

	if ctx.GlobalIsSet("http-port") || c.HttpServerPort == 0 {
		c.HttpServerPort = ctx.GlobalInt("http-port")
	}

	if ctx.GlobalIsSet("http-mode") || c.HttpServerMode == "" {
		c.HttpServerMode = ctx.GlobalString("http-mode")
	}

	return nil
}
