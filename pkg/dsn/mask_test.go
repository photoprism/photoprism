package dsn

import (
	"testing"
)

//nolint:gosec // G101: DSN masking tests intentionally use inline credential samples.
func TestMask(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{name: "Empty", in: "", out: ""},
		{name: "NoPassword", in: "user@tcp(localhost:3306)/db", out: "user@tcp(localhost:3306)/db"},
		{name: "WithPassword", in: "user:secret@tcp(localhost:3306)/db", out: "user:***@tcp(localhost:3306)/db"},
		{name: "URIStyle", in: "mysql://user:secret@localhost:3306/db?parseTime=true", out: "mysql://user:***@localhost:3306/db?parseTime=true"},
		{name: "FallbackWhenNoNeedle", in: "user:secret@tcp(localhost:3306)/db?password=secret", out: "user:***@tcp(localhost:3306)/db?password=secret"},
		{name: "UnixSocket", in: "user:secret@unix(/var/run/mysql.sock)/db", out: "user:***@unix(/var/run/mysql.sock)/db"},
		{name: "NoPasswordQuery", in: "user@tcp(localhost:3306)/db?password=secret", out: "user@tcp(localhost:3306)/db?password=secret"},
		{name: "Postgres", in: "user=alice password=s3cr3t dbname=app", out: "user=alice password=*** dbname=app"},
		{name: "PostgresQuoted", in: "user=alice password=\"s ec ret\" dbname=app", out: "user=alice password=\"***\" dbname=app"},
		{name: "PostgresQueryPassword", in: "postgres://u@h/db?password=secret", out: "postgres://u@h/db?password=***"},
		{name: "PostgresKeyValueNoUser", in: "host=h password=secret dbname=x", out: "host=h password=*** dbname=x"},
		{name: "PostgresKeyValueURIOption", in: "host=h user=u password=secret dbname=db options=http://x:y@z", out: "host=h user=u password=*** dbname=db options=http://x:y@z"},
		{name: "PostgresKeyValueURIPassword", in: "host=h user=u password=a://b:c@d dbname=db", out: "host=h user=u password=*** dbname=db"},
		{name: "MySQLUserWithEquals", in: "a=b:secret@tcp(h)/db", out: "a=b:***@tcp(h)/db"},
		{name: "PostgresKeyValueEscapedSpace", in: `user=u password=a\ b dbname=x`, out: "user=u password=*** dbname=x"},
		{name: "PostgresKeyValueAfterQuote", in: "host='h'password=Q19 dbname=x", out: "host='h'password=*** dbname=x"},
		{name: "PostgresKeyValueVerticalTab", in: "host=h\vpassword=S18", out: "host=h\vpassword=***"},
		{name: "MySQLURISlash", in: "mysql://u:p@ss/x@tcp(h:3306)/db", out: "mysql://u:***@tcp(h:3306)/db"},
		{name: "PostgresURIEncoded", in: "postgres://alice:p%40ss%2Fw@db:5432/app?sslmode=disable", out: "postgres://alice:***@db:5432/app?sslmode=disable"},
		{name: "PostgresURISocket", in: "postgresql://alice:secret@/app?host=%2Fvar%2Frun%2Fpostgresql", out: "postgresql://alice:***@/app?host=%2Fvar%2Frun%2Fpostgresql"},
		{name: "PostgresSingleQuoted", in: "password='secret' user=alice dbname=app", out: "password='***' user=alice dbname=app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Mask(tt.in); got != tt.out {
				t.Fatalf("Mask(%q) = %q, want %q", tt.in, got, tt.out)
			}
		})
	}
}
