ALTER TABLE files MODIFY IF EXISTS file_diff BIGINT DEFAULT -1;
ALTER TABLE files MODIFY IF EXISTS file_chroma SMALLINT(6) DEFAULT -1;
ALTER TABLE auth_sessions MODIFY IF EXISTS refresh_token VARBINARY(2048) DEFAULT '';
ALTER TABLE auth_sessions MODIFY IF EXISTS id_token VARBINARY(2048) DEFAULT '';
DROP INDEX IF EXISTS idx_accounts_deleted_at ON services;
DROP INDEX IF EXISTS idx_files_file_main_color ON files;
