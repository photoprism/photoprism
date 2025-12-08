package dsn

import (
	"testing"
)

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
