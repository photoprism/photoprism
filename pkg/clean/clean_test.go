package clean

import (
	"path/filepath"
)

// ProjectRoot references the project root directory for use in tests.
var ProjectRoot = func() string { dir, _ := filepath.Abs("../../"); return dir }()
