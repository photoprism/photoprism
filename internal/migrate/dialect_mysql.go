package migrate

// Generated code, do not edit.

var DialectMySQL = Migrations{
	{
		ID:         "20211121-094727",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS uix_places_place_label ON places;"},
	},
	{
		ID:         "20211124-120008",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS idx_places_place_label ON places;", "DROP INDEX IF EXISTS uix_places_label ON places;"},
	},
	{
		ID:         "20220329-030000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY file_projection VARBINARY(64) NULL;", "ALTER TABLE files MODIFY file_color_profile VARBINARY(64) NULL;"},
	},
	{
		ID:         "20220329-040000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS idx_albums_album_filter ON albums;", "ALTER TABLE albums MODIFY album_filter VARBINARY(2048) DEFAULT '';", "CREATE OR REPLACE INDEX idx_albums_album_filter ON albums (album_filter(512));"},
	},
	{
		ID:         "20220329-050000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE photos MODIFY photo_description VARCHAR(4096);"},
	},
	{
		ID:         "20220329-060000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE albums MODIFY album_caption VARCHAR(1024);", "ALTER TABLE albums MODIFY album_description VARCHAR(2048);", "ALTER TABLE albums MODIFY album_notes VARCHAR(1024);", "ALTER TABLE cameras MODIFY camera_description VARCHAR(2048);", "ALTER TABLE cameras MODIFY camera_notes VARCHAR(1024);", "ALTER TABLE countries MODIFY country_description VARCHAR(2048);", "ALTER TABLE countries MODIFY country_notes VARCHAR(1024);", "ALTER TABLE details MODIFY keywords VARCHAR(2048);", "ALTER TABLE details MODIFY notes VARCHAR(2048);", "ALTER TABLE details MODIFY subject VARCHAR(1024);", "ALTER TABLE details MODIFY artist VARCHAR(1024);", "ALTER TABLE details MODIFY copyright VARCHAR(1024);", "ALTER TABLE details MODIFY license VARCHAR(1024);", "ALTER TABLE folders MODIFY folder_description VARCHAR(2048);", "ALTER TABLE labels MODIFY label_description VARCHAR(2048);", "ALTER TABLE labels MODIFY label_notes VARCHAR(1024);", "ALTER TABLE lenses MODIFY lens_description VARCHAR(2048);", "ALTER TABLE lenses MODIFY lens_notes VARCHAR(1024);", "ALTER TABLE subjects MODIFY subj_bio VARCHAR(2048);", "ALTER TABLE subjects MODIFY subj_notes VARCHAR(1024);"},
	},
	{
		ID:         "20220329-061000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"CREATE OR REPLACE INDEX idx_files_photo_id ON files (photo_id, file_primary);"},
	},
	{
		ID:         "20220329-070000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY COLUMN IF EXISTS photo_taken_at DATETIME AFTER photo_uid;", "ALTER TABLE files ADD COLUMN IF NOT EXISTS photo_taken_at DATETIME AFTER photo_uid;"},
	},
	{
		ID:         "20220329-071000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"UPDATE files f JOIN photos p ON p.id = f.photo_id SET f.photo_taken_at = p.taken_at_local;"},
	},
	{
		ID:         "20220329-080000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY IF EXISTS media_id VARBINARY(32) AFTER photo_taken_at;", "ALTER TABLE files ADD IF NOT EXISTS media_id VARBINARY(32) AFTER photo_taken_at;"},
	},
	{
		ID:         "20220329-081000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"CREATE OR REPLACE UNIQUE INDEX idx_files_search_media ON files (media_id);"},
	},
	{
		ID:         "20220329-083000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"UPDATE files SET media_id = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN CONCAT((10000000000 - photo_id), '-', 1 + file_sidecar - file_primary, '-', file_uid) END;"},
	},
	{
		ID:         "20220329-090000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY IF EXISTS time_index VARBINARY(64) AFTER photo_taken_at;", "ALTER TABLE files ADD IF NOT EXISTS time_index VARBINARY(64) AFTER photo_taken_at;"},
	},
	{
		ID:         "20220329-091000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"CREATE OR REPLACE UNIQUE INDEX idx_files_search_timeline ON files (time_index);"},
	},
	{
		ID:         "20220329-093000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"UPDATE files SET time_index = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN CONCAT(100000000000000 - CAST(photo_taken_at AS UNSIGNED), '-', media_id) END;"},
	},
	{
		ID:         "20220421-200000",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"CREATE OR REPLACE INDEX idx_files_missing_root ON files (file_missing, file_root);"},
	},
	{
		ID:         "20220521-000001",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE photos MODIFY photo_color SMALLINT DEFAULT -1;"},
	},
	{
		ID:         "20220521-000002",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY file_diff INTEGER DEFAULT -1;"},
	},
	{
		ID:         "20220521-000003",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY file_chroma SMALLINT DEFAULT -1;"},
	},
	{
		ID:         "20220927-000100",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY time_index VARBINARY(64);"},
	},
	{
		ID:         "20221002-000100",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE links DROP COLUMN IF EXISTS can_edit;", "ALTER TABLE links DROP COLUMN IF EXISTS can_comment;"},
	},
	{
		ID:         "20221015-100000",
		Dialect:    "mysql",
		Stage:      "pre",
		Statements: []string{"RENAME TABLE IF EXISTS `accounts` TO `services`;"},
	},
	{
		ID:         "20221015-100100",
		Dialect:    "mysql",
		Stage:      "pre",
		Statements: []string{"ALTER IGNORE TABLE files_sync CHANGE account_id service_id INT UNSIGNED NOT NULL;", "ALTER IGNORE TABLE files_share CHANGE account_id service_id INT UNSIGNED NOT NULL;"},
	},
	{
		ID:         "20230102-000001",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE albums MODIFY IF EXISTS album_path VARCHAR(1024);"},
	},
	{
		ID:         "20230211-000001",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files MODIFY IF EXISTS file_colors VARBINARY(18);", "ALTER TABLE files MODIFY IF EXISTS File_luminance VARBINARY(18);"},
	},
	{
		ID:         "20230309-000001",
		Dialect:    "mysql",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users SET auth_provider = 'local' WHERE id = 1;", "UPDATE auth_users SET auth_provider = 'none' WHERE id = -1;", "UPDATE auth_users SET auth_provider = 'token' WHERE id = -2;", "UPDATE auth_users SET auth_provider = 'default' WHERE auth_provider = '' OR auth_provider = 'password' OR auth_provider IS NULL;"},
	},
}
