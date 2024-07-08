package i18n

const (
	ErrUnexpected Message = iota + 1
	ErrBadRequest
	ErrSaveFailed
	ErrDeleteFailed
	ErrAlreadyExists
	ErrNotFound
	ErrFileNotFound
	ErrFileTooLarge
	ErrUnsupported
	ErrUnsupportedType
	ErrUnsupportedFormat
	ErrOriginalsEmpty
	ErrSelectionNotFound
	ErrEntityNotFound
	ErrAccountNotFound
	ErrUserNotFound
	ErrLabelNotFound
	ErrAlbumNotFound
	ErrSubjectNotFound
	ErrPersonNotFound
	ErrFaceNotFound
	ErrPublic
	ErrReadOnly
	ErrUnauthorized
	ErrForbidden
	ErrOffensiveUpload
	ErrUploadFailed
	ErrNoItemsSelected
	ErrCreateFile
	ErrCreateFolder
	ErrConnectionFailed
	ErrPasscodeRequired
	ErrInvalidPasscode
	ErrInvalidPassword
	ErrFeatureDisabled
	ErrNoLabelsSelected
	ErrNoAlbumsSelected
	ErrNoFilesForDownload
	ErrZipFailed
	ErrInvalidCredentials
	ErrInvalidLink
	ErrInvalidName
	ErrBusy
	ErrWakeupInterval
	ErrAccountConnect
	ErrTooManyRequests

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
	MsgSubjectSaved
	MsgSubjectDeleted
	MsgPersonSaved
	MsgPersonDeleted
	MsgFileUploaded
	MsgFilesUploadedIn
	MsgProcessingUpload
	MsgUploadProcessed
	MsgSelectionApproved
	MsgSelectionArchived
	MsgSelectionRestored
	MsgSelectionProtected
	MsgAlbumsDeleted
	MsgZipCreatedIn
	MsgPermanentlyDeleted
	MsgRestored
	MsgVerified
	MsgActivated
)

var Messages = MessageMap{
	// Error messages:
	ErrUnexpected:         gettext("Something went wrong, try again"),
	ErrBadRequest:         gettext("Unable to do that"),
	ErrSaveFailed:         gettext("Changes could not be saved"),
	ErrDeleteFailed:       gettext("Could not be deleted"),
	ErrAlreadyExists:      gettext("%s already exists"),
	ErrNotFound:           gettext("Not found"),
	ErrFileNotFound:       gettext("File not found"),
	ErrFileTooLarge:       gettext("File too large"),
	ErrUnsupported:        gettext("Unsupported"),
	ErrUnsupportedType:    gettext("Unsupported type"),
	ErrUnsupportedFormat:  gettext("Unsupported format"),
	ErrOriginalsEmpty:     gettext("Originals folder is empty"),
	ErrSelectionNotFound:  gettext("Selection not found"),
	ErrEntityNotFound:     gettext("Entity not found"),
	ErrAccountNotFound:    gettext("Account not found"),
	ErrUserNotFound:       gettext("User not found"),
	ErrLabelNotFound:      gettext("Label not found"),
	ErrAlbumNotFound:      gettext("Album not found"),
	ErrSubjectNotFound:    gettext("Subject not found"),
	ErrPersonNotFound:     gettext("Person not found"),
	ErrFaceNotFound:       gettext("Face not found"),
	ErrPublic:             gettext("Not available in public mode"),
	ErrReadOnly:           gettext("Not available in read-only mode"),
	ErrUnauthorized:       gettext("Please log in to your account"),
	ErrForbidden:          gettext("Permission denied"),
	ErrOffensiveUpload:    gettext("Upload might be offensive"),
	ErrUploadFailed:       gettext("Upload failed"),
	ErrNoItemsSelected:    gettext("No items selected"),
	ErrCreateFile:         gettext("Failed creating file, please check permissions"),
	ErrCreateFolder:       gettext("Failed creating folder, please check permissions"),
	ErrConnectionFailed:   gettext("Could not connect, please try again"),
	ErrPasscodeRequired:   gettext("Enter verification code"),
	ErrInvalidPasscode:    gettext("Invalid verification code, please try again"),
	ErrInvalidPassword:    gettext("Invalid password, please try again"),
	ErrFeatureDisabled:    gettext("Feature disabled"),
	ErrNoLabelsSelected:   gettext("No labels selected"),
	ErrNoAlbumsSelected:   gettext("No albums selected"),
	ErrNoFilesForDownload: gettext("No files available for download"),
	ErrZipFailed:          gettext("Failed to create zip file"),
	ErrInvalidCredentials: gettext("Invalid credentials"),
	ErrInvalidLink:        gettext("Invalid link"),
	ErrInvalidName:        gettext("Invalid name"),
	ErrBusy:               gettext("Busy, please try again later"),
	ErrWakeupInterval:     gettext("The wakeup interval is %s, but must be 1h or less"),
	ErrAccountConnect:     gettext("Your account could not be connected"),
	ErrTooManyRequests:    gettext("Too many requests"),

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
	MsgSubjectSaved:          gettext("Subject saved"),
	MsgSubjectDeleted:        gettext("Subject deleted"),
	MsgPersonSaved:           gettext("Person saved"),
	MsgPersonDeleted:         gettext("Person deleted"),
	MsgFileUploaded:          gettext("File uploaded"),
	MsgFilesUploadedIn:       gettext("%d files uploaded in %d s"),
	MsgProcessingUpload:      gettext("Processing upload..."),
	MsgUploadProcessed:       gettext("Upload has been processed"),
	MsgSelectionApproved:     gettext("Selection approved"),
	MsgSelectionArchived:     gettext("Selection archived"),
	MsgSelectionRestored:     gettext("Selection restored"),
	MsgSelectionProtected:    gettext("Selection marked as private"),
	MsgAlbumsDeleted:         gettext("Albums deleted"),
	MsgZipCreatedIn:          gettext("Zip created in %d s"),
	MsgPermanentlyDeleted:    gettext("Permanently deleted"),
	MsgRestored:              gettext("%s has been restored"),
	MsgVerified:              gettext("Successfully verified"),
	MsgActivated:             gettext("Successfully activated"),
}
