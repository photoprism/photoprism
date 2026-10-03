package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRegisterDatabase_Server checks the server address built from a registration response.
func TestRegisterDatabase_Server(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, tc := range []struct {
			host   string
			port   int
			server string
		}{
			{"mariadb", 4001, "mariadb:4001"},
			{"mariadb", 0, "mariadb"},
			{"::1", 3306, "[::1]:3306"},
			{"", 3306, ":3306"},
			{"", 0, ""},
		} {
			server, ok := RegisterDatabase{Host: tc.host, Port: tc.port}.Server()
			assert.True(t, ok, "%s:%d", tc.host, tc.port)
			assert.Equal(t, tc.server, server)
		}
	})
	t.Run("Unusable", func(t *testing.T) {
		for _, host := range []string{"/run/mysqld/mysqld.sock", "-x", "db/x"} {
			_, ok := RegisterDatabase{Host: host, Port: 3306}.Server()
			assert.False(t, ok, host)
		}
	})
}
