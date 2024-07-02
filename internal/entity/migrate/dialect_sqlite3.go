package migrate

// Generated code, do not edit.

var DialectSQLite3 = Migrations{
	{
		ID:         "20211121-094727",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS idx_places_place_label;"},
	},
	{
		ID:         "20211124-120008",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS uix_places_place_label;", "DROP INDEX IF EXISTS uix_places_label;"},
	},
	{
		ID:         "20220329-040000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS idx_albums_album_filter;"},
	},
	{
		ID:         "20220329-050000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_albums_album_filter ON albums (album_filter);"},
	},
	{
		ID:         "20220329-061000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_files_photo_id ON files (photo_id, file_primary);"},
	},
	{
		ID:         "20220329-071000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE files SET photo_taken_at = (SELECT taken_at_local FROM photos WHERE photos.id = photo_id) WHERE photo_id IS NOT NULL;"},
	},
	{
		ID:         "20220329-081000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE UNIQUE INDEX IF NOT EXISTS idx_files_search_media ON files (media_id);"},
	},
	{
		ID:         "20220329-083000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE files SET media_id = CASE WHEN photo_id IS NOT NULL AND file_missing = 0 AND deleted_at IS NULL THEN ((10000000000 - photo_id) || '-' || (1 + file_sidecar - file_primary) || '-' || file_uid) END WHERE 1;"},
	},
	{
		ID:         "20220329-091000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE UNIQUE INDEX IF NOT EXISTS idx_files_search_timeline ON files (time_index);"},
	},
	{
		ID:         "20220329-093000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE files SET time_index = CASE WHEN media_id IS NOT NULL AND photo_taken_at IS NOT NULL THEN ((100000000000000 - strftime('%Y%m%d%H%M%S', photo_taken_at)) || '-' || media_id) ELSE NULL END WHERE photo_id IS NOT NULL;"},
	},
	{
		ID:         "20220421-200000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_files_missing_root ON files (file_missing, file_root);"},
	},
	{
		ID:         "20221015-100000",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE accounts RENAME TO services;"},
	},
	{
		ID:         "20221015-100100",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE files_sync RENAME COLUMN account_id TO service_id;", "ALTER TABLE files_share RENAME COLUMN account_id TO service_id;"},
	},
	{
		ID:         "20230309-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users SET auth_provider = 'local' WHERE id = 1;", "UPDATE auth_users SET auth_provider = 'none' WHERE id = -1;", "UPDATE auth_users SET auth_provider = 'token' WHERE id = -2;", "UPDATE auth_users SET auth_provider = 'default' WHERE auth_provider = '' OR auth_provider = 'password' OR auth_provider IS NULL;"},
	},
	{
		ID:         "20230313-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users SET user_role = 'contributor' WHERE user_role = 'uploader';", "UPDATE auth_sessions SET auth_provider = 'link' WHERE auth_provider = 'token';"},
	},
	{
		ID:         "20240112-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DELETE FROM auth_sessions;"},
	},
}
