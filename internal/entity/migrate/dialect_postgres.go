package migrate

// Generated code, do not edit.

// DialectPostgres is the migrations for the DBMS Postgres
var DialectPostgres = Migrations{
	{
		ID:         "20241202-000001",
		Dialect:    "postgres",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users_details SET birth_year = -1 WHERE birth_year >= 0 AND birth_year < 1000 OR birth_year < -1 OR birth_year IS NULL;", "UPDATE auth_users_details SET birth_month = -1 WHERE birth_month = 0 OR birth_month < -1 OR birth_month > 12 OR birth_month IS NULL;", "UPDATE auth_users_details SET birth_day = -1 WHERE birth_day = 0 OR birth_day < -1 OR birth_day > 31 OR birth_day IS NULL;", "UPDATE auth_users_details SET user_country = 'zz' WHERE user_country = '' OR user_country IS NULL;"},
	},
	{
		ID:         "20250117-000001",
		Dialect:    "postgres",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE IF EXISTS photos RENAME COLUMN photo_description TO photo_caption;", "ALTER TABLE IF EXISTS photos RENAME COLUMN description_src TO caption_src;"},
	},
	{
		ID:         "20250416-000001",
		Dialect:    "postgres",
		Stage:      "main",
		Statements: []string{"UPDATE photos SET time_zone = 'Local' WHERE time_zone = '' OR time_zone IS NULL;"},
	},
	{
		ID:         "20250819-000001",
		Dialect:    "postgres",
		Stage:      "post",
		Statements: []string{"CREATE COLLATION IF NOT EXISTS caseinsensitive (provider = icu, locale = 'und', deterministic = false);"},
	},
	{
		ID:         "20260615-000001",
		Dialect:    "postgres",
		Stage:      "main",
		Statements: []string{"ALTER TABLE files ALTER COLUMN file_diff SET DEFAULT -1;", "ALTER TABLE files ALTER COLUMN file_chroma SET DEFAULT -1;"},
	},
	{
		ID:         "20260711-000001",
		Dialect:    "postgres",
		Stage:      "post",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_albums_album_filter ON albums (album_filter);", "CREATE INDEX IF NOT EXISTS idx_albums_album_path ON albums (album_path);"},
	},
	{
		ID:         "20260822-000001",
		Dialect:    "postgres",
		Stage:      "post",
		Statements: []string{"CREATE OR REPLACE FUNCTION public.safe_make_date(year bigint, month bigint, day bigint)\nRETURNS date \nLANGUAGE plpgsql IMMUTABLE AS \n$$\nDECLARE\n\tsyear smallint;\n\tsmonth smallint;\n\tsday smallint;\nBEGIN\n\tBEGIN\n\t\tsyear := year::smallint;\n\t\tsmonth := month::smallint;\n\t\tsday := day::smallint;\n    \tRETURN make_date(syear, smonth, sday);\n\tEXCEPTION\n\t    WHEN others THEN \n    \t    RETURN null;\n\tEND;\nEND;\n$$;"},
	},
}
