package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

const (
	DbTiDB  = "internal"
	DbMySQL = "mysql"
)

// Params provides a struct in which application configuration is stored.
// Application code must use functions to get config values, for two reasons:
//
// 1. Some values are computed and we don't want to leak implementation details (aims at reducing refactoring overhead).
//
// 2. Paths might actually be dynamic later (if we build a multi-user version).
//
// See https://github.com/photoprism/photoprism/issues/50#issuecomment-433856358
type Params struct {
	Name               string
	Version            string
	Copyright          string
	Debug              bool   `yaml:"debug" flag:"debug"`
	ReadOnly           bool   `yaml:"read-only" flag:"read-only"`
	LogLevel           string `yaml:"log-level" flag:"log-level"`
	ConfigFile         string
	AssetsPath         string `yaml:"assets-path" flag:"assets-path"`
	CachePath          string `yaml:"cache-path" flag:"cache-path"`
	OriginalsPath      string `yaml:"originals-path" flag:"originals-path"`
	ImportPath         string `yaml:"import-path" flag:"import-path"`
	ExportPath         string `yaml:"export-path" flag:"export-path"`
	SqlServerHost      string `yaml:"sql-host" flag:"sql-host"`
	SqlServerPort      uint   `yaml:"sql-port" flag:"sql-port"`
	SqlServerPath      string `yaml:"sql-path" flag:"sql-path"`
	SqlServerPassword  string `yaml:"sql-password" flag:"sql-password"`
	HttpServerHost     string `yaml:"http-host" flag:"http-host"`
	HttpServerPort     int    `yaml:"http-port" flag:"http-port"`
	HttpServerMode     string `yaml:"http-mode" flag:"http-mode"`
	HttpServerPassword string `yaml:"http-password" flag:"http-password"`
	DarktableCli       string `yaml:"darktable-cli" flag:"darktable-cli"`
	DatabaseDriver     string `yaml:"database-driver" flag:"database-driver"`
	DatabaseDsn        string `yaml:"database-dsn" flag:"database-dsn"`
}

// NewParams() creates a new configuration entity by using two methods:
//
// 1. SetValuesFromFile: This will initialize values from a yaml config file.
//
// 2. SetValuesFromCliContext: Which comes after SetValuesFromFile and overrides
//    any previous values giving an option two override file configs through the CLI.
func NewParams(ctx *cli.Context) *Params {
	c := &Params{}

	c.Name = ctx.App.Name
	c.Copyright = ctx.App.Copyright
	c.Version = ctx.App.Version

	if err := c.SetValuesFromFile(util.ExpandedFilename(ctx.GlobalString("config-file"))); err != nil {
		log.Debug(err)
	}

	if err := c.SetValuesFromCliContext(ctx); err != nil {
		log.Error(err)
	}

	c.expandFilenames()

	return c
}

func (c *Params) expandFilenames() {
	c.AssetsPath = util.ExpandedFilename(c.AssetsPath)
	c.CachePath = util.ExpandedFilename(c.CachePath)
	c.OriginalsPath = util.ExpandedFilename(c.OriginalsPath)
	c.ImportPath = util.ExpandedFilename(c.ImportPath)
	c.ExportPath = util.ExpandedFilename(c.ExportPath)
	c.DarktableCli = util.ExpandedFilename(c.DarktableCli)
	c.SqlServerPath = util.ExpandedFilename(c.SqlServerPath)
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (c *Params) SetValuesFromFile(fileName string) error {
	if !util.Exists(fileName) {
		return errors.New(fmt.Sprintf("config file not found: \"%s\"", fileName))
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(yamlConfig, c)
}

// SetValuesFromCliContext uses values from the CLI to setup configuration overrides
// for the entity.
func (c *Params) SetValuesFromCliContext(ctx *cli.Context) error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through all config fields
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagValue := v.Type().Field(i).Tag.Get("flag")

		// Automatically assign values to fields with "flag" tag
		if tagValue != "" {
			switch t := fieldValue.Interface().(type) {
			case int, int64:
				// Only if explicitly set or current value is empty (use default)
				if ctx.GlobalIsSet(tagValue) || fieldValue.Int() == 0 {
					f := ctx.GlobalInt64(tagValue)
					fieldValue.SetInt(f)
				}
			case uint, uint64:
				// Only if explicitly set or current value is empty (use default)
				if ctx.GlobalIsSet(tagValue) || fieldValue.Uint() == 0 {
					f := ctx.GlobalUint64(tagValue)
					fieldValue.SetUint(f)
				}
			case string:
				// Only if explicitly set or current value is empty (use default)
				if ctx.GlobalIsSet(tagValue) || fieldValue.String() == "" {
					f := ctx.GlobalString(tagValue)
					fieldValue.SetString(f)
				}
			case bool:
				if ctx.GlobalIsSet(tagValue) {
					f := ctx.GlobalBool(tagValue)
					fieldValue.SetBool(f)
				}
			default:
				log.Warnf("can't assign value of type %s from cli flag %s", t, tagValue)
			}
		}
	}

	return nil
}
