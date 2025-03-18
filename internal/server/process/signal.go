package process

import (
	"os"
	"syscall"
)

// Signal channel for initiating the server shutdown.
var Signal = make(chan os.Signal)

// Restart gracefully restarts the server.
//
// Note that this requires an entrypoint script or other process to
// spawns a new instance when the server exists with status code 1.
func Restart() {
	Signal <- syscall.SIGUSR1
}

// Shutdown gracefully stops the server.
func Shutdown() {
	Signal <- syscall.SIGTERM
}
