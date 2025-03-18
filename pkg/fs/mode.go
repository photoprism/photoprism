package fs

import (
	"os"
	"strconv"
)

// File and directory permissions.
var (
	ModeDir    os.FileMode = 0o777
	ModeSocket os.FileMode = 0o666
	ModeFile   os.FileMode = 0o666
	ModeBackup os.FileMode = 0o600
)

// ParseMode parses and returns a filesystem permission mode,
// or the specified default mode if it could not be parsed.
func ParseMode(s string, defaultMode os.FileMode) os.FileMode {
	if s == "" {
		return defaultMode
	}
	mode, err := strconv.ParseUint(s, 8, 32)

	if err != nil || mode <= 0 {
		return defaultMode
	}

	return os.FileMode(mode)
}
