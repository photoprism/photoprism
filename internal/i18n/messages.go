package i18n

const (
	ErrUnexpected Message = iota + 1
	ErrBadRequest
	ErrSaveFailed
	ErrDeleteFailed
	ErrAlreadyExists
	ErrNotFound
	ErrFileNotFound
	ErrSelectionNotFound
	ErrEntityNotFound
	ErrAccountNotFound
	ErrUserNotFound
	ErrLabelNotFound
	ErrAlbumNotFound
	ErrPublic
	ErrReadOnly
	ErrUnauthorized
	ErrOffensiveUpload
	ErrNoItemsSelected
	ErrCreateFile
	ErrCreateFolder
	ErrConnectionFailed
	ErrInvalidPassword
	ErrFeatureDisabled
	ErrNoLabelsSelected
	ErrNoAlbumsSelected
	ErrNoFilesForDownload
	ErrZipFailed
	ErrInvalidCredentials
	ErrInvalidLink

	MsgChangesSaved
	MsgAlbumCreated
	MsgAlbumSaved
	MsgAlbumDeleted
	MsgAlbumCloned
	MsgFileUnstacked
	MsgFileDeleted
	MsgSelectionAddedTo
	MsgEntryAddedTo
	MsgEntriesAddedTo
	MsgEntryRemovedFrom
	MsgEntriesRemovedFrom
	MsgAccountCreated
	MsgAccountSaved
	MsgAccountDeleted
	MsgSettingsSaved
	MsgPasswordChanged
	MsgImportCompletedIn
	MsgImportCanceled
	MsgIndexingCompletedIn
	MsgIndexingOriginals
	MsgIndexingFiles
	MsgIndexingCanceled
	MsgRemovedFilesAndPhotos
	MsgMovingFilesFrom
	MsgCopyingFilesFrom
	MsgLabelsDeleted
	MsgLabelSaved
	MsgFilesUploadedIn
	MsgSelectionApproved
	MsgSelectionArchived
	MsgSelectionRestored
	MsgSelectionProtected
	MsgAlbumsDeleted
	MsgZipCreatedIn
	MsgPermanentlyDeleted
)

var Messages = MessageMap{
	// Error messages:
	ErrUnexpected:         gettext("Unexpected error, please try again"),
	ErrBadRequest:         gettext("Invalid request"),
	ErrSaveFailed:         gettext("Changes could not be saved"),
	ErrDeleteFailed:       gettext("Could not be deleted"),
	ErrAlreadyExists:      gettext("%s already exists"),
	ErrNotFound:           gettext("Not found on server, deleted?"),
	ErrFileNotFound:       gettext("File not found"),
	ErrSelectionNotFound:  gettext("Selection not found"),
	ErrEntityNotFound:     gettext("Not found on server, deleted?"),
	ErrAccountNotFound:    gettext("Account not found"),
	ErrUserNotFound:       gettext("User not found"),
	ErrLabelNotFound:      gettext("Label not found"),
	ErrAlbumNotFound:      gettext("Album not found"),
	ErrPublic:             gettext("Not available in public mode"),
	ErrReadOnly:           gettext("not available in read-only mode"),
	ErrUnauthorized:       gettext("Please log in and try again"),
	ErrOffensiveUpload:    gettext("Upload might be offensive"),
	ErrNoItemsSelected:    gettext("No items selected"),
	ErrCreateFile:         gettext("Failed creating file, please check permissions"),
	ErrCreateFolder:       gettext("Failed creating folder, please check permissions"),
	ErrConnectionFailed:   gettext("Could not connect, please try again"),
	ErrInvalidPassword:    gettext("Invalid password, please try again"),
	ErrFeatureDisabled:    gettext("Feature disabled"),
	ErrNoLabelsSelected:   gettext("No labels selected"),
	ErrNoAlbumsSelected:   gettext("No albums selected"),
	ErrNoFilesForDownload: gettext("No files available for download"),
	ErrZipFailed:          gettext("Failed to create zip file"),
	ErrInvalidCredentials: gettext("Invalid credentials"),
	ErrInvalidLink:        gettext("Invalid link"),

	// Info and confirmation messages:
	MsgChangesSaved:          gettext("Changes successfully saved"),
	MsgAlbumCreated:          gettext("Album created"),
	MsgAlbumSaved:            gettext("Album saved"),
	MsgAlbumDeleted:          gettext("Album %s deleted"),
	MsgAlbumCloned:           gettext("Album contents cloned"),
	MsgFileUnstacked:         gettext("File removed from stack"),
	MsgFileDeleted:           gettext("File deleted"),
	MsgSelectionAddedTo:      gettext("Selection added to %s"),
	MsgEntryAddedTo:          gettext("One entry added to %s"),
	MsgEntriesAddedTo:        gettext("%d entries added to %s"),
	MsgEntryRemovedFrom:      gettext("One entry removed from %s"),
	MsgEntriesRemovedFrom:    gettext("%d entries removed from %s"),
	MsgAccountCreated:        gettext("Account created"),
	MsgAccountSaved:          gettext("Account saved"),
	MsgAccountDeleted:        gettext("Account deleted"),
	MsgSettingsSaved:         gettext("Settings saved"),
	MsgPasswordChanged:       gettext("Password changed"),
	MsgImportCompletedIn:     gettext("Import completed in %d s"),
	MsgImportCanceled:        gettext("Import canceled"),
	MsgIndexingCompletedIn:   gettext("Indexing completed in %d s"),
	MsgIndexingOriginals:     gettext("Indexing originals..."),
	MsgIndexingFiles:         gettext("Indexing files in %s"),
	MsgIndexingCanceled:      gettext("Indexing canceled"),
	MsgRemovedFilesAndPhotos: gettext("Removed %d files and %d photos"),
	MsgMovingFilesFrom:       gettext("Moving files from %s"),
	MsgCopyingFilesFrom:      gettext("Copying files from %s"),
	MsgLabelsDeleted:         gettext("Labels deleted"),
	MsgLabelSaved:            gettext("Label saved"),
	MsgFilesUploadedIn:       gettext("%d files uploaded in %d s"),
	MsgSelectionApproved:     gettext("Selection approved"),
	MsgSelectionArchived:     gettext("Selection archived"),
	MsgSelectionRestored:     gettext("Selection restored"),
	MsgSelectionProtected:    gettext("Selection marked as private"),
	MsgAlbumsDeleted:         gettext("Albums deleted"),
	MsgZipCreatedIn:          gettext("Zip created in %d s"),
	MsgPermanentlyDeleted:    gettext("Permanently deleted"),
}
