package mutex

// Activities that can be started and stopped.
var (
	MainWorker   = Activity{}
	SyncWorker   = Activity{}
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

// WorkersRunning checks if a worker is currently running.
func WorkersRunning() bool {
	return MainWorker.Running() || SyncWorker.Running() || ShareWorker.Running() || MetaWorker.Running() || FacesWorker.Running()
}
