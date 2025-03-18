package process

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"
)

// ID is the process ID under which the server is running.
var ID = os.Getpid()

// WritePID writes the process ID to a file, if specified.
func WritePID(fileName string) error {
	if fileName == "" {
		return nil
	}

	if pidDir := filepath.Dir(fileName); !fs.Writable(pidDir) {
		return fmt.Errorf("%s is not writable", pidDir)
	} else if err := fs.WriteString(fileName, fmt.Sprintf("%d", ID)); err != nil {
		return err
	}

	return nil
}
