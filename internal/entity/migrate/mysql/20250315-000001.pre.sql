ALTER TABLE auth_users_settings CHANGE COLUMN IF EXISTS default_page ui_start_page VARCHAR(64) DEFAULT 'default';
