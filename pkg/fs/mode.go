package fs

import "os"

// File and directory permissions.
var (
	ModeDir    os.FileMode = 0o777
	ModeFile   os.FileMode = 0o666
	ModeBackup os.FileMode = 0o600
)
