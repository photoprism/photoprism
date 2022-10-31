package fs

import "os"

var (
	ModeDir  os.FileMode = 0o777
	ModeFile os.FileMode = 0o666
)
