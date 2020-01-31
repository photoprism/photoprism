package photoprism

type ImportOptions struct {
	Path                   string
	Move                   bool
	RemoveDotFiles         bool
	RemoveExistingFiles    bool
	RemoveEmptyDirectories bool
}

// ImportOptionsCopy returns import options for copying files to originals (read-only).
func ImportOptionsCopy(path string) ImportOptions {
	result := ImportOptions{
		Path:                   path,
		Move:                   false,
		RemoveDotFiles:         false,
		RemoveExistingFiles:    false,
		RemoveEmptyDirectories: false,
	}

	return result
}

// IndexOptionsMove returns import options for moving files to originals (modifies import directory).
func ImportOptionsMove(path string) ImportOptions {
	result := ImportOptions{
		Path:                   path,
		Move:                   true,
		RemoveDotFiles:         true,
		RemoveExistingFiles:    true,
		RemoveEmptyDirectories: true,
	}

	return result
}
