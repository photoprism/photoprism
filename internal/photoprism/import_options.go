package photoprism

import "github.com/photoprism/photoprism/internal/entity"

// ImportOptions represents file import options.
type ImportOptions struct {
	UID                    string
	Action                 string
	Albums                 []string
	Path                   string
	Move                   bool
	NonBlocking            bool
	DestFolder             string
	RemoveDotFiles         bool
	RemoveExistingFiles    bool
	RemoveEmptyDirectories bool
}

// SetUser sets the user who performs the import operation.
func (o *ImportOptions) SetUser(user *entity.User) *ImportOptions {
	if o != nil && user != nil {
		o.UID = user.GetUID()
	}

	return o
}

// ImportOptionsCopy returns import options for copying files to originals (read-only).
func ImportOptionsCopy(importPath, destFolder string) ImportOptions {
	result := ImportOptions{
		UID:                    entity.Admin.GetUID(),
		Action:                 ActionImport,
		Path:                   importPath,
		Move:                   false,
		NonBlocking:            false,
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
		UID:                    entity.Admin.GetUID(),
		Action:                 ActionImport,
		Path:                   importPath,
		Move:                   true,
		NonBlocking:            false,
		DestFolder:             destFolder,
		RemoveDotFiles:         true,
		RemoveExistingFiles:    true,
		RemoveEmptyDirectories: true,
	}

	return result
}

// ImportOptionsUpload returns options for importing user uploads.
func ImportOptionsUpload(uploadPath, destFolder string) ImportOptions {
	result := ImportOptions{
		UID:                    entity.Admin.GetUID(),
		Action:                 ActionUpload,
		Path:                   uploadPath,
		Move:                   true,
		NonBlocking:            true,
		DestFolder:             destFolder,
		RemoveDotFiles:         true,
		RemoveExistingFiles:    true,
		RemoveEmptyDirectories: true,
	}

	return result
}
