package photoprism

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/frontend"
	"github.com/sirupsen/logrus"
)

// Config interface implemented in context (cli) and test packages
type Config interface {
	Debug() bool
	LogLevel() logrus.Level
	Db() *gorm.DB

	CreateDirectories() error
	MigrateDb()

	ClientConfig() frontend.Config
	ConfigFile() string

	AppName() string
	AppVersion() string
	AppCopyright() string

	SqlServerHost() string
	SqlServerPort() uint
	SqlServerPath() string
	SqlServerPassword() string

	HttpServerHost() string
	HttpServerPort() int
	HttpServerMode() string
	HttpServerPassword() string
	HttpTemplatesPath() string
	HttpFaviconsPath() string
	HttpPublicPath() string
	HttpPublicBuildPath() string

	DatabaseDriver() string
	DatabaseDsn() string

	AssetsPath() string
	ServerPath() string
	OriginalsPath() string
	ImportPath() string
	ExportPath() string
	CachePath() string
	ThumbnailsPath() string
	TensorFlowModelPath() string

	DarktableCli() string
}
