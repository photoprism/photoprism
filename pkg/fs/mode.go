package fs

import (
	"os"
	"strconv"
)

// File and directory permissions. Umask restricts
// further; these are not the effective permissions.
var (
	ModeDir        os.FileMode = 0o777 // Default directory mode (POSIX).
	ModeSocket     os.FileMode = 0o666
	ModeFile       os.FileMode = 0o666 // Default modes for regular files.
	ModeConfigFile os.FileMode = 0o664
	ModeSecretFile os.FileMode = 0o600
	ModeBackupFile os.FileMode = 0o600
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
