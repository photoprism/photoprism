package entity

import "github.com/jinzhu/gorm"

// Deprecated represents a list of deprecated database tables.
type Deprecated []string

// Drop drops all deprecated tables.
func (list Deprecated) Drop(db *gorm.DB) {
	for _, tableName := range list {
		if err := db.DropTableIfExists(tableName).Error; err != nil {
			log.Debugf("drop %s: %s", tableName, err)
		}
	}
}

// DeprecatedTables lists deprecated development database tables, so that they can be removed.
var DeprecatedTables = Deprecated{
	"subjects_dev1",
	"subjects_dev2",
	"subjects_dev3",
	"subjects_dev4",
	"subjects_dev5",
	"subjects_dev6",
	"subjects_dev7",
	"subjects_dev8",
	"subjects_dev9",
	"subjects_dev10",
	"markers_dev1",
	"markers_dev2",
	"markers_dev3",
	"markers_dev4",
	"markers_dev5",
	"markers_dev6",
	"markers_dev7",
	"markers_dev8",
	"markers_dev9",
	"markers_dev10",
	"faces_dev1",
	"faces_dev2",
	"faces_dev3",
	"faces_dev4",
	"faces_dev5",
	"faces_dev6",
	"faces_dev7",
	"faces_dev8",
	"faces_dev9",
	"faces_dev10",
	"reactions_dev",
	"albums_users_dev",
	"photos_users_dev",
	"auth_sessions_dev",
	"auth_shares_dev",
	"auth_users_details_dev",
	"auth_users_dev",
	"auth_users_logins_dev",
	"auth_users_settings_dev",
	"auth_tokens_dev",
	"addresses",
}
