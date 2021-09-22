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

type IndexResult struct {
	Status   IndexStatus
	Err      error
	FileID   uint
	FileUID  string
	PhotoID  uint
	PhotoUID string
}

func (r IndexResult) String() string {
	return string(r.Status)
}

func (r IndexResult) Failed() bool {
	return r.Err != nil
}

func (r IndexResult) Success() bool {
	return r.Err == nil && (r.FileID > 0 || r.Stacked() || r.Skipped() || r.Archived())
}

func (r IndexResult) Indexed() bool {
	return r.Status == IndexAdded || r.Status == IndexUpdated || r.Status == IndexStacked
}

func (r IndexResult) Stacked() bool {
	return r.Status == IndexStacked
}

func (r IndexResult) Skipped() bool {
	return r.Status == IndexSkipped
}

func (r IndexResult) Archived() bool {
	return r.Status == IndexArchived
}
