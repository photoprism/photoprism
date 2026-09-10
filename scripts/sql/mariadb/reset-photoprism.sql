-- 	# Warning:  This will reset the photoprism database which is the default database, not a testing database.
DROP DATABASE IF EXISTS photoprism;
CREATE DATABASE IF NOT EXISTS photoprism;

FLUSH PRIVILEGES;
