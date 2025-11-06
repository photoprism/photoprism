package dsn

import "github.com/photoprism/photoprism/pkg/enum"

// Params maps required DSN parameters by driver type.
var Params = Values{
	enum.MySQL:    "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	enum.MariaDB:  "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	enum.Postgres: "sslmode=disable TimeZone=UTC",
	enum.SQLite3:  "_busy_timeout=5000",
}
