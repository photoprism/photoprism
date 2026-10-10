package cluster

import (
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/http/dns"
)

// Server returns the database server address of a registration response, the host with the port if one
// is set, and whether an instance can connect to it. Without a host or port, it returns "" and true.
func (d RegisterDatabase) Server() (server string, ok bool) {
	server = d.Host

	if d.Port > 0 {
		server = dns.JoinHostPort(d.Host, d.Port)
	}

	return server, server == "" || dsn.ValidServer(server)
}
