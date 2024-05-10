package mutex

// Activities that can be started and stopped.
var (
	MainWorker   = Activity{}
	SyncWorker   = Activity{}
	BackupWorker = Activity{}
	ShareWorker  = Activity{}
	MetaWorker   = Activity{}
	FacesWorker  = Activity{}
	UpdatePeople = Activity{}
)

// CancelAll requests to stop all activities.
func CancelAll() {
	UpdatePeople.Cancel()
	MainWorker.Cancel()
	SyncWorker.Cancel()
	ShareWorker.Cancel()
	MetaWorker.Cancel()
	FacesWorker.Cancel()
}

// IndexWorkersRunning checks if a worker is currently running.
func IndexWorkersRunning() bool {
	return MainWorker.Running() || SyncWorker.Running() || ShareWorker.Running() || MetaWorker.Running() || FacesWorker.Running()
}
