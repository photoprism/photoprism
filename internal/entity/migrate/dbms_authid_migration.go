package migrate

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

// ConvertDBMSAuthIDDataTypes applies the data type conversion for auth_id columns for SQLite
// Conversion is from VARBINARY(255) DEFAULT ” to TEXT NOT NULL COLLATE BINARY DEFAULT ”
// This can't be done in a script as the order of columns is not known, and is determined by the age of the database
func ConvertDBMSAuthIDDataTypes(db *gorm.DB) (err error) {
	switch db.Dialect().GetName() {
	case SQLite3:
		// These create statements will get out of date, but that is ok, as the main migrate path will add any missing columns/indexes in later.
		authSessionsCreate := `CREATE TABLE "auth_sessions" ("id" VARBINARY(2048),"user_uid" VARBINARY(42) DEFAULT '',"user_name" varchar(200),"client_uid" VARBINARY(42) DEFAULT '',"client_name" varchar(200) DEFAULT '',"client_ip" varchar(64),"auth_provider" VARBINARY(128) DEFAULT '',"auth_method" VARBINARY(128) DEFAULT '',"auth_issuer" VARBINARY(255) DEFAULT '',"auth_id" TEXT NOT NULL COLLATE BINARY DEFAULT '',"auth_scope" varchar(1024) DEFAULT '',"grant_type" VARBINARY(64) DEFAULT '',"last_active" bigint,"sess_expires" bigint,"sess_timeout" bigint,"preview_token" VARBINARY(64) DEFAULT '',"download_token" VARBINARY(64) DEFAULT '',"access_token" VARBINARY(4096) DEFAULT '',"refresh_token" VARBINARY(2048) DEFAULT '',"id_token" VARBINARY(2048) DEFAULT '',"user_agent" varchar(512),"data_json" VARBINARY(4096),"ref_id" VARBINARY(16) DEFAULT '',"login_ip" varchar(64),"login_at" datetime,"created_at" datetime,"updated_at" datetime , PRIMARY KEY ("id"))`
		authUsersCreate := `CREATE TABLE "auth_users" ("id" integer primary key autoincrement,"user_uuid" VARBINARY(64),"user_uid" VARBINARY(42),"auth_provider" VARBINARY(128) DEFAULT '',"auth_method" VARBINARY(128) DEFAULT '',"auth_issuer" VARBINARY(255) DEFAULT '',"auth_id" TEXT NOT NULL COLLATE BINARY DEFAULT '',"user_name" varchar(200),"display_name" varchar(200),"user_email" varchar(255),"backup_email" varchar(255),"user_role" varchar(64) DEFAULT '',"user_scope" varchar(1024) DEFAULT '*',"user_attr" varchar(1024) DEFAULT '',"super_admin" bool,"can_login" bool,"login_at" datetime,"expires_at" datetime,"webdav" bool,"base_path" VARBINARY(1024),"upload_path" VARBINARY(1024),"can_invite" bool,"invite_token" VARBINARY(64),"invited_by" varchar(64),"verify_token" VARBINARY(64),"verified_at" datetime,"consent_at" datetime,"born_at" datetime,"reset_token" VARBINARY(64),"preview_token" VARBINARY(64),"download_token" VARBINARY(64),"thumb" VARBINARY(128) DEFAULT '',"thumb_src" VARBINARY(8) DEFAULT '',"ref_id" VARBINARY(16),"created_at" datetime,"updated_at" datetime,"deleted_at" datetime )`

		type resultIndex struct {
			Name string
		}

		type pragmaTable struct {
			Cid       int
			Name      string
			Type      string
			Notnull   int
			DfltValue string
			Pk        int
		}

		// Start a transaction
		tx := db.Begin()

		if tx.Error != nil {
			return fmt.Errorf("migrate: error creating transaction %w", tx.Error)
		}

		defer func() {
			if err == nil {
				if txErr := tx.Commit().Error; txErr != nil {
					log.Warningf("migrate: commit failure for DBMS AuthID Data Types: %w", txErr)
				} else {
					log.Debug("migrate: committed DBMS AuthID Data Types")
				}
			} else {
				if txErr := tx.Rollback().Error; txErr != nil {
					log.Warningf("migrate: rollback failure for DBMS AuthID Data Types: %w", txErr)
				} else {
					log.Warning("migrate: rolled back DBMS AuthID Data Types")
				}
			}
		}()

		if !tx.HasTable("auth_sessions") {
			if err = tx.Exec(authSessionsCreate).Error; err != nil {
				return fmt.Errorf("migrate: error creating auth_sessions %w", err)
			}
		} else {
			// Data Migration here, by rename, create new, data transfer, drop indexes
			if err = tx.Exec(`ALTER TABLE "auth_sessions" RENAME TO "migrate_auth_sessions"`).Error; err != nil {
				return fmt.Errorf("migrate: error renaming auth_sessions %w", err)
			}
			if err = tx.Exec(authSessionsCreate).Error; err != nil {
				return fmt.Errorf("migrate: error creating auth_sessions %w", err)
			}

			// Get the columns of both old and new table, and find the columns that are in the old and new table
			var oldPragmaColumns []pragmaTable
			var newPragmaColumns []pragmaTable
			oldColumns := make(map[string]bool)

			if err = tx.Raw("PRAGMA table_info(migrate_auth_sessions)").Scan(&oldPragmaColumns).Error; err != nil {
				return fmt.Errorf("migrate: error getting column list for migrate_auth_sessions with %w", err)
			}
			for _, pragma := range oldPragmaColumns {
				oldColumns[pragma.Name] = false
			}

			if err = tx.Raw("PRAGMA table_info(auth_sessions)").Scan(&newPragmaColumns).Error; err != nil {
				return fmt.Errorf("migrate: error getting column list for auth_sessions with %w", err)
			}
			for _, pragma := range newPragmaColumns {
				if _, present := oldColumns[pragma.Name]; present {
					oldColumns[pragma.Name] = true
				}
			}
			// Build the select into statement
			var columns []string
			for key, value := range oldColumns {
				if value {
					columns = append(columns, key)
				}
			}

			populateStmt := fmt.Sprintf("INSERT INTO auth_sessions (%s) SELECT %s FROM migrate_auth_sessions", strings.Join(columns, ", "), strings.Join(columns, ", "))

			if err = tx.Exec(populateStmt).Error; err != nil {
				return fmt.Errorf("migrate: error migrating with stmt %s with %w", populateStmt, err)
			}

			var indexes []resultIndex
			if err = tx.Raw("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL", "migrate_auth_sessions").Scan(&indexes).Error; err != nil {
				return fmt.Errorf("migrate: error getting index list %w", err)
			}
			for _, index := range indexes {
				dropStatement := fmt.Sprintf(`DROP INDEX IF EXISTS "%s"`, index.Name)
				if err = tx.Exec(dropStatement).Error; err != nil {
					return fmt.Errorf("migrate: error dropping index %s was %w", index.Name, err)
				}
			}

			if err = tx.Exec("DROP TABLE migrate_auth_sessions").Error; err != nil {
				return fmt.Errorf("migrate: error dropping table migrate_auth_sessions with %w", err)
			}
		}
		if !tx.HasTable("auth_users") {
			if err = tx.Exec(authUsersCreate).Error; err != nil {
				return fmt.Errorf("migrate: error creating auth_users %w", err)
			}
		} else {
			// Data Migration here, by rename, create new, data transfer, drop indexes
			if err = tx.Exec(`ALTER TABLE "auth_users" RENAME TO "migrate_auth_users"`).Error; err != nil {
				return fmt.Errorf("migrate: error renaming auth_users %w", err)
			}
			if err = tx.Exec(authUsersCreate).Error; err != nil {
				return fmt.Errorf("migrate: error creating auth_users %w", err)
			}

			// Get the columns of both old and new table, and find the columns that are in the old and new table
			var oldPragmaColumns []pragmaTable
			var newPragmaColumns []pragmaTable
			oldColumns := make(map[string]bool)

			if err = tx.Raw("PRAGMA table_info(migrate_auth_users)").Scan(&oldPragmaColumns).Error; err != nil {
				return fmt.Errorf("migrate: error getting column list for migrate_auth_users with %w", err)
			}
			for _, pragma := range oldPragmaColumns {
				oldColumns[pragma.Name] = false
			}

			if err = tx.Raw("PRAGMA table_info(auth_users)").Scan(&newPragmaColumns).Error; err != nil {
				return fmt.Errorf("migrate: error getting column list for auth_users with %w", err)
			}
			for _, pragma := range newPragmaColumns {
				if _, present := oldColumns[pragma.Name]; present {
					oldColumns[pragma.Name] = true
				}
			}
			// Build the select into statement
			var columns []string
			for key, value := range oldColumns {
				if value {
					columns = append(columns, key)
				}
			}

			populateStmt := fmt.Sprintf("INSERT INTO auth_users (%s) SELECT %s FROM migrate_auth_users", strings.Join(columns, ", "), strings.Join(columns, ", "))

			if err = tx.Exec(populateStmt).Error; err != nil {
				return fmt.Errorf("migrate: error migrating with stmt %s with %w", populateStmt, err)
			}

			var indexes []resultIndex
			if err = tx.Raw("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL", "migrate_auth_users").Scan(&indexes).Error; err != nil {
				return fmt.Errorf("migrate: error getting index list %w", err)
			}
			for _, index := range indexes {
				dropStatement := fmt.Sprintf(`DROP INDEX IF EXISTS "%s"`, index.Name)
				if err = tx.Exec(dropStatement).Error; err != nil {
					return fmt.Errorf("migrate: error dropping index %s was %w", index.Name, err)
				}
			}

			if err = tx.Exec("DROP TABLE migrate_auth_users").Error; err != nil {
				return fmt.Errorf("migrate: error dropping table migrate_auth_users with %w", err)
			}
		}
	// case MySQL: // MySQL
	// Nothing required for Gorm V1.  Statements left in comments for Gorm V2 implementation.
	// 	// These create statements will get out of date, but that is ok, as the main migrate path will add any missing columns/indexes in later.
	// 	authSessionsCreate := "CREATE TABLE `auth_sessions` (`id` VARBINARY(2048),`user_uid` VARBINARY(42) DEFAULT '',`user_name` varchar(200),`client_uid` VARBINARY(42) DEFAULT '',`client_name` varchar(200) DEFAULT '',`client_ip` varchar(64),`auth_provider` VARBINARY(128) DEFAULT '',`auth_method` VARBINARY(128) DEFAULT '',`auth_issuer` VARBINARY(255) DEFAULT '',`auth_id` VARBINARY(255) DEFAULT '',`auth_scope` varchar(1024) DEFAULT '',`grant_type` VARBINARY(64) DEFAULT '',`last_active` bigint,`sess_expires` bigint,`sess_timeout` bigint,`preview_token` VARBINARY(64) DEFAULT '',`download_token` VARBINARY(64) DEFAULT '',`access_token` VARBINARY(4096) DEFAULT '',`refresh_token` VARBINARY(2048) DEFAULT '',`id_token` VARBINARY(2048) DEFAULT '',`user_agent` varchar(512),`data_json` VARBINARY(4096),`ref_id` VARBINARY(16) DEFAULT '',`login_ip` varchar(64),`login_at` DATETIME NULL,`created_at` DATETIME NULL,`updated_at` DATETIME NULL , PRIMARY KEY (`id`))"
	// 	authUsersCreate := "CREATE TABLE `auth_users` (`id` int AUTO_INCREMENT,`user_uuid` VARBINARY(64),`user_uid` VARBINARY(42),`auth_provider` VARBINARY(128) DEFAULT '',`auth_method` VARBINARY(128) DEFAULT '',`auth_issuer` VARBINARY(255) DEFAULT '',`auth_id` VARBINARY(255) DEFAULT '',`user_name` varchar(200),`display_name` varchar(200),`user_email` varchar(255),`backup_email` varchar(255),`user_role` varchar(64) DEFAULT '',`user_scope` varchar(1024) DEFAULT '*',`user_attr` varchar(1024) DEFAULT '',`super_admin` boolean,`can_login` boolean,`login_at` DATETIME NULL,`expires_at` DATETIME NULL,`webdav` boolean,`base_path` VARBINARY(1024),`upload_path` VARBINARY(1024),`can_invite` boolean,`invite_token` VARBINARY(64),`invited_by` varchar(64),`verify_token` VARBINARY(64),`verified_at` DATETIME NULL,`consent_at` DATETIME NULL,`born_at` DATETIME NULL,`reset_token` VARBINARY(64),`preview_token` VARBINARY(64),`download_token` VARBINARY(64),`thumb` VARBINARY(128) DEFAULT '',`thumb_src` VARBINARY(8) DEFAULT '',`ref_id` VARBINARY(16),`created_at` DATETIME NULL,`updated_at` DATETIME NULL,`deleted_at` DATETIME NULL , PRIMARY KEY (`id`))"
	// 	if !db.HasTable("auth_sessions") {
	// 		if err := db.Exec(authSessionsCreate).Error; err != nil {
	// 			return fmt.Errorf("migrate: error creating auth_sessions %w", err)
	// 		}
	// 	}
	// 	if !db.HasTable("auth_users") {
	// 		if err := db.Exec(authUsersCreate).Error; err != nil {
	// 			return fmt.Errorf("migrate: error creating auth_users %w", err)
	// 		}
	// 	}
	//	// There are no migration needs for MariaDB as the structure is not being manipulated.
	// case Postgres:
	// Nothing required for Gorm V1
	default:
	}
	return nil
}
