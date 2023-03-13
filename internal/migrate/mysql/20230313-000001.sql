UPDATE auth_users SET user_role = 'contributor' WHERE user_role = 'uploader';
UPDATE auth_sessions SET auth_provider = 'link' WHERE auth_provider = 'token';