ALTER IGNORE TABLE files_sync CHANGE account_id service_id INT UNSIGNED NOT NULL;
ALTER IGNORE TABLE files_share CHANGE account_id service_id INT UNSIGNED NOT NULL;