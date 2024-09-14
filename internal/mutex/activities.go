package mutex

// Activities that can be started and stopped.
var (
	IndexWorker  = Activity{}
	SyncWorker   = Activity{}
	BackupWorker = Activity{}
	ShareWorker  = Activity{}
	MetaWorker   = Activity{}
	FacesWorker  = Activity{}
	UpdatePeople = Activity{}
)

// CancelAll requests to stop all activities.
func CancelAll() {
	IndexWorker.Cancel()
	SyncWorker.Cancel()
	BackupWorker.Cancel()
	ShareWorker.Cancel()
	MetaWorker.Cancel()
	FacesWorker.Cancel()
	UpdatePeople.Cancel()
}

// WorkersRunning checks if a worker is currently running.
func WorkersRunning() bool {
	return IndexWorker.Running() || SyncWorker.Running() || BackupWorker.Running() || ShareWorker.Running() || MetaWorker.Running() || FacesWorker.Running()
}
