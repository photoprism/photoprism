package photoprism

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Config interface {
	CreateDirectories() error
	MigrateDb()

	GetDb() *gorm.DB
	GetClientConfig() map[string]interface{}

	GetConfigFile() string
	GetAppName() string
	GetAppVersion() string
	GetAppCopyright() string
	IsDebug() bool
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

