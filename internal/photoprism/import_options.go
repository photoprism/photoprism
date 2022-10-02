package photoprism

// ImportOptions represents file import options.
type ImportOptions struct {
	Albums                 []string
	Path                   string
	Move                   bool
	UserUID                string
	DestFolder             string
	RemoveDotFiles         bool
	RemoveExistingFiles    bool
	RemoveEmptyDirectories bool
}

// ImportOptionsCopy returns import options for copying files to originals (read-only).
func ImportOptionsCopy(importPath, destFolder string) ImportOptions {
	result := ImportOptions{
		Path:                   importPath,
		Move:                   false,
		DestFolder:             destFolder,
		RemoveDotFiles:         false,
		RemoveExistingFiles:    false,
		RemoveEmptyDirectories: false,
	}

	return result
}

// ImportOptionsMove returns import options for moving files to originals (modifies import directory).
func ImportOptionsMove(importPath, destFolder string) ImportOptions {
	result := ImportOptions{
		Path:                   importPath,
		Move:                   true,
		DestFolder:             destFolder,
		RemoveDotFiles:         true,
		RemoveExistingFiles:    true,
		RemoveEmptyDirectories: true,
	}

	return result
}
