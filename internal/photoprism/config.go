package photoprism

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/frontend"
)

// Todo: Remove Get prefix, see https://golang.org/doc/effective_go.html#Getters
type Config interface {
	CreateDirectories() error
	MigrateDb()

	Db() *gorm.DB
	ClientConfig() frontend.Config

	ConfigFile() string
	AppName() string
	AppVersion() string
	AppCopyright() string
	IsDebug() bool
	DbServerIP() string
	DbServerPort() uint
	GetServerIP() string
	GetServerPort() int
	GetServerMode() string
	GetDatabaseDriver() string
	GetDatabaseDsn() string
	GetOriginalsPath() string
	GetImportPath() string
	GetExportPath() string
	GetDarktableCli() string
	GetCachePath() string
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
