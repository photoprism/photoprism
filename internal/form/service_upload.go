package form

// SyncUpload defines payload for syncing uploads to a remote service.
type SyncUpload struct {
	Selection Selection `json:"selection"`
	Folder    string    `json:"folder"`
}
