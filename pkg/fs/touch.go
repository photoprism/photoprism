package fs

import (
	"os"
	"strconv"
	"time"
)

// Touch creates a file with the specified name that contains the current Unix timestamp.
// If the file path does not exist or it cannot be written, an error is returned.
func Touch(fileName string) error {
	return os.WriteFile(fileName, []byte(strconv.FormatInt(time.Now().Unix(), 10)), ModeFile)
}
