package server

import (
	"github.com/photoprism/photoprism/internal/server/process"
)

// Fail logs an error and then initiates a server shutdown.
func Fail(err string, params ...interface{}) {
	if err != "" {
		log.Errorf(err, params...)
	}

	process.Shutdown()
}
