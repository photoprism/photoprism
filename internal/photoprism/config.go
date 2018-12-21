package photoprism

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/frontend"
)

// Todo: Remove Get prefix, see https://golang.org/doc/effective_go.html#Getters
type Config interface {
	Debug() bool
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

	HttpServerHost() string
	HttpServerPort() int
	HttpServerMode() string
	ServerPath() string
	HttpTemplatesPath() string
	HttpFaviconsPath() string
	HttpPublicPath() string
	HttpPublicBuildPath() string

	DatabaseDriver() string
	DatabaseDsn() string

	AssetsPath() string
	OriginalsPath() string
	ImportPath() string
	ExportPath() string
	CachePath() string
	ThumbnailsPath() string
	TensorFlowModelPath() string

	DarktableCli() string
}
