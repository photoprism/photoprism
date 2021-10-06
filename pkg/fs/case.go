package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// CaseInsensitive tests if a storage path is case-insensitive.
func CaseInsensitive(storagePath string) (result bool, err error) {
	tmpName := filepath.Join(storagePath, "caseTest.tmp")

	if err := os.WriteFile(tmpName, []byte("{}"), 0666); err != nil {
		return false, fmt.Errorf("%s not writable", filepath.Base(storagePath))
	}

	defer os.Remove(tmpName)

	result = FileExists(filepath.Join(storagePath, "CASETEST.TMP"))

	return result, err
}

// IgnoreCase enables the case-insensitive mode.
func IgnoreCase() {
	ignoreCase = true
	TypeExt = FileExt.TypeExt()
}
