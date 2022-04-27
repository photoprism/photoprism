package migrate

// Generated code, do not edit.

var DialectMySQL = Migrations{
	{
		ID:         "20211121-094727",
		Dialect:    "mysql",
		Statements: []string{"DROP INDEX IF EXISTS uix_places_place_label ON places;"},
	},
	{
		ID:         "20211124-120008",
		Dialect:    "mysql",
		Statements: []string{"DROP INDEX IF EXISTS idx_places_place_label ON places;", "DROP INDEX IF EXISTS uix_places_label ON places;"},
	},
	{
		ID:         "20220329-030000",
		Dialect:    "mysql",
		Statements: []string{"ALTER TABLE files MODIFY file_projection VARBINARY(64) NULL;", "ALTER TABLE files MODIFY file_color_profile VARBINARY(64) NULL;"},
	},
	{
		ID:         "20220329-040000",
		Dialect:    "mysql",
		Statements: []string{"DROP INDEX IF EXISTS idx_albums_album_filter ON albums;", "ALTER TABLE albums MODIFY album_filter VARBINARY(2048) DEFAULT '';", "CREATE OR REPLACE INDEX idx_albums_album_filter ON albums (album_filter(512));"},
	},
	{
		ID:         "20220329-050000",
		Dialect:    "mysql",
		Statements: []string{"UPDATE photos SET photo_description = SUBSTR(photo_description, 0, 4096) WHERE 1;", "ALTER TABLE photos MODIFY photo_description VARCHAR(4096);"},
	},
	{
		ID:         "20220329-060000",
		Dialect:    "mysql",
		Statements: []string{"ALTER TABLE accounts MODIFY acc_url VARCHAR(255);", "ALTER TABLE addresses MODIFY address_notes VARCHAR(1024);", "ALTER TABLE albums MODIFY album_caption VARCHAR(1024);", "ALTER TABLE albums MODIFY album_description VARCHAR(2048);", "ALTER TABLE albums MODIFY album_notes VARCHAR(1024);", "ALTER TABLE cameras MODIFY camera_description VARCHAR(2048);", "ALTER TABLE cameras MODIFY camera_notes VARCHAR(1024);", "ALTER TABLE countries MODIFY country_description VARCHAR(2048);", "ALTER TABLE countries MODIFY country_notes VARCHAR(1024);", "UPDATE details SET keywords = SUBSTR(keywords, 0, 2048), notes = SUBSTR(notes, 0, 2048) WHERE 1;", "ALTER TABLE details MODIFY keywords VARCHAR(2048);", "ALTER TABLE details MODIFY notes VARCHAR(2048);", "ALTER TABLE details MODIFY subject VARCHAR(1024);", "ALTER TABLE details MODIFY artist VARCHAR(1024);", "ALTER TABLE details MODIFY copyright VARCHAR(1024);", "ALTER TABLE details MODIFY license VARCHAR(1024);", "UPDATE folders SET folder_description = SUBSTR(folder_description, 0, 2048) WHERE 1;", "ALTER TABLE folders MODIFY folder_description VARCHAR(2048);", "ALTER TABLE labels MODIFY label_description VARCHAR(2048);", "ALTER TABLE labels MODIFY label_notes VARCHAR(1024);", "ALTER TABLE lenses MODIFY lens_description VARCHAR(2048);", "ALTER TABLE lenses MODIFY lens_notes VARCHAR(1024);", "ALTER TABLE subjects MODIFY subj_bio VARCHAR(2048);", "ALTER TABLE subjects MODIFY subj_notes VARCHAR(1024);"},
	},
	{
		ID:         "20220329-061000",
		Dialect:    "mysql",
		Statements: []string{"CREATE OR REPLACE INDEX idx_files_photo_id ON files (photo_id, file_primary);"},
	},
	{
		ID:         "20220329-070000",
		Dialect:    "mysql",
		Statements: []string{"ALTER TABLE files MODIFY COLUMN IF EXISTS photo_taken_at DATETIME AFTER photo_uid;", "ALTER TABLE files ADD COLUMN IF NOT EXISTS photo_taken_at DATETIME AFTER photo_uid;"},
	},
	{
		ID:         "20220329-071000",
		Dialect:    "mysql",
		Statements: []string{"UPDATE files f JOIN photos p ON p.id = f.photo_id SET f.photo_taken_at = p.taken_at_local;"},
	},
	{
		ID:         "20220329-080000",
		Dialect:    "mysql",
		Statements: []string{"ALTER TABLE files MODIFY IF EXISTS media_id VARBINARY(32) AFTER photo_taken_at;", "ALTER TABLE files ADD IF NOT EXISTS media_id VARBINARY(32) AFTER photo_taken_at;"},
	},
	{
		ID:         "20220329-081000",
		Dialect:    "mysql",
		Statements: []string{"CREATE OR REPLACE UNIQUE INDEX idx_files_search_media ON files (media_id);"},
	},
	{
		ID:         "20220329-083000",
		Dialect:    "mysql",
		Statements: []string{"UPDATE files SET media_id = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN CONCAT((10000000000 - photo_id), '-', 1 + file_sidecar - file_primary, '-', file_uid) END;"},
	},
	{
		ID:         "20220329-090000",
		Dialect:    "mysql",
		Statements: []string{"ALTER TABLE files MODIFY IF EXISTS time_index VARBINARY(48) AFTER photo_taken_at;", "ALTER TABLE files ADD IF NOT EXISTS time_index VARBINARY(48) AFTER photo_taken_at;"},
	},
	{
		ID:         "20220329-091000",
		Dialect:    "mysql",
		Statements: []string{"CREATE OR REPLACE UNIQUE INDEX idx_files_search_timeline ON files (time_index);"},
	},
	{
		ID:         "20220329-093000",
		Dialect:    "mysql",
		Statements: []string{"UPDATE files SET time_index = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN CONCAT(100000000000000 - CAST(photo_taken_at AS UNSIGNED), '-', media_id) END;"},
	},
	{
		ID:         "20220421-200000",
		Dialect:    "mysql",
		Statements: []string{"CREATE OR REPLACE INDEX idx_files_missing_root ON files (file_missing, file_root);"},
	},
}
