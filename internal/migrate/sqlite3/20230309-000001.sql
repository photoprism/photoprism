UPDATE auth_users SET auth_provider = 'local' WHERE id = 1;
UPDATE auth_users SET auth_provider = 'none' WHERE id = -1;
UPDATE auth_users SET auth_provider = 'token' WHERE id = -2;
UPDATE auth_users SET auth_provider = 'default' WHERE auth_provider = '' OR auth_provider = 'password' OR auth_provider IS NULL;