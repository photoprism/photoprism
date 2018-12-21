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

	HttpServerHost() string
	HttpServerPort() int
	HttpServerMode() string

	DatabaseDriver() string
	DatabaseDsn() string

	OriginalsPath() string
	ImportPath() string
	ExportPath() string
	CachePath() string

	DarktableCli() string
	GetThumbnailsPath() string
	GetAssetsPath() string
	GetTensorFlowModelPath() string
	GetDatabasePath() string
	GetServerAssetsPath() string
	GetTemplatesPath() string
	GetFaviconsPath() string
	GetPublicPath() string
	GetPublicBuildPath() string
}
