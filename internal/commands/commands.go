/*
Package commands contains commands and flags used by the photoprism application.

Additional information concerning the command-line interface can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Commands
*/
package commands

import (
	"os"
	"syscall"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/sevlyar/go-daemon"
)

var log = event.Log

// chilAlreadyRunning test if a .pid file at filePath is a running proccess.
// it returns the pid value and the running status (true or false).
func childAlreadyRunning(filePath string) (pid int, running bool) {
	if !fs.FileExists(filePath) {
		return pid, false
	}

	pid, err := daemon.ReadPidFile(filePath)
	if err != nil {
		return pid, false
	}

	process, err := os.FindProcess(int(pid))
	if err != nil {
		return pid, false
	}

	return pid, process.Signal(syscall.Signal(0)) == nil
}
