package photoprism

const (
	IndexUpdated   IndexStatus = "updated"
	IndexAdded     IndexStatus = "added"
	IndexStacked   IndexStatus = "stacked"
	IndexSkipped   IndexStatus = "skipped"
	IndexDuplicate IndexStatus = "skipped duplicate"
	IndexArchived  IndexStatus = "skipped archived"
	IndexFailed    IndexStatus = "failed"
)

type IndexStatus string

// IndexResult represents a media file indexing result.
type IndexResult struct {
	Status   IndexStatus
	Err      error
	FileID   uint
	FileUID  string
	PhotoID  uint
	PhotoUID string
}

// String returns the indexing result as string.
func (r IndexResult) String() string {
	return string(r.Status)
}

// Success checks whether a media file was successfully indexed or skipped.
func (r IndexResult) Success() bool {
	return !r.Failed() && (r.FileID > 0 || r.Stacked() || r.Skipped() || r.Archived())
}

// Failed checks if indexing has failed.
func (r IndexResult) Failed() bool {
	return r.Err != nil && r.Status == IndexFailed
}

// Indexed checks whether a media file was successfully indexed.
func (r IndexResult) Indexed() bool {
	return r.Status == IndexAdded || r.Status == IndexUpdated || r.Status == IndexStacked
}

// Stacked checks whether a media file was stacked while indexing.
func (r IndexResult) Stacked() bool {
	return r.Status == IndexStacked
}

// Skipped checks whether a media file was skipped while indexing.
func (r IndexResult) Skipped() bool {
	return r.Status == IndexSkipped
}

// Archived checks whether a media file was skipped because it is archived.
func (r IndexResult) Archived() bool {
	return r.Status == IndexArchived
}

// FileError checks if there is a file error and returns it.
func (r IndexResult) FileError() (string, error) {
	if r.Failed() && r.FileUID != "" {
		return r.FileUID, r.Err
	} else {
		return "", nil
	}
}
