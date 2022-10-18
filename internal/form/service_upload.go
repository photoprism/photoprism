package form

type SyncUpload struct {
	Selection Selection `json:"selection"`
	Folder    string    `json:"folder"`
}
