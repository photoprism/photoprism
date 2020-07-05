package i18n

const (
	ErrUnexpected Message = iota + 1
	ErrBadRequest
	ErrSaveFailed
	ErrAlreadyExists
	ErrEntityNotFound
	ErrAccountNotFound
	ErrAlbumNotFound
	ErrReadOnly
	ErrUnauthorized
	ErrUploadNSFW
	ErrNoItemsSelected
	ErrCreateFile
	ErrCreateFolder
	ErrConnectionFailed

	MsgChangesSaved
	MsgAlbumCreated
	MsgAlbumSaved
	MsgAlbumDeleted
	MsgAlbumCloned
	MsgFileUngrouped
	MsgSelectionAddedTo
	MsgEntryAddedTo
	MsgEntriesAddedTo
	MsgEntryRemovedFrom
	MsgEntriesRemovedFrom
	MsgAccountCreated
	MsgAccountSaved
	MsgAccountDeleted
	MsgSettingsSaved
)

var MsgEnglish = MessageMap{
	ErrUnexpected:      "Unexpected error, please try again",
	ErrBadRequest:      "Invalid request",
	ErrSaveFailed:      "Changes could not be saved",
	ErrAlreadyExists:   "%s already exists",
	ErrEntityNotFound:  "Unknown entity",
	ErrAccountNotFound: "Unknown account",
	ErrAlbumNotFound:   "Album not found",
	ErrReadOnly:        "not available in read-only mode",
	ErrUnauthorized:    "Please log in and try again",
	ErrUploadNSFW:      "Upload might be offensive",
	ErrNoItemsSelected: "No items selected",
	ErrCreateFile:      "Failed creating file, please check permissions",
	ErrCreateFolder:    "Failed creating folder, please check permissions",
	ErrConnectionFailed: "Could not connect, please try again",

	MsgChangesSaved:       "Changes successfully saved",
	MsgAlbumCreated:       "Album created",
	MsgAlbumSaved:         "Album saved",
	MsgAlbumDeleted:       "Album %s deleted",
	MsgAlbumCloned:        "Album contents cloned",
	MsgFileUngrouped:      "File successfully ungrouped",
	MsgSelectionAddedTo:   "Selection added to %s",
	MsgEntryAddedTo:       "One entry added to %s",
	MsgEntriesAddedTo:     "%d entries added to %s",
	MsgEntryRemovedFrom:   "One entry removed from %s",
	MsgEntriesRemovedFrom: "%d entries removed from %s",
	MsgAccountCreated:     "Account created",
	MsgAccountSaved:       "Account saved",
	MsgAccountDeleted:     "Account deleted",
	MsgSettingsSaved:      "Settings saved",
}
