UPDATE auth_users_details SET birth_year = -1 WHERE birth_year >= 0 AND birth_year < 1000 OR birth_year < -1 OR birth_year IS NULL;
UPDATE auth_users_details SET birth_month = -1 WHERE birth_month = 0 OR birth_month < -1 OR birth_month > 12 OR birth_month IS NULL;
UPDATE auth_users_details SET birth_day = -1 WHERE birth_day = 0 OR birth_day < -1 OR birth_day > 31 OR birth_day IS NULL;
UPDATE auth_users_details SET user_country = 'zz' WHERE user_country = '' OR user_country IS NULL;