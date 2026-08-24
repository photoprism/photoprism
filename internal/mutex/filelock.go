package mutex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
)

// FileLockMaxAge is how long a file lock stays valid without being renewed.
//
// A lock that never expires wedges an instance whenever its holder is killed, and only a human
// who knows the file exists can free it again. This bounds that to one interval, and the holder
// renews well inside it for as long as it is alive.
const FileLockMaxAge = 5 * time.Minute

// fileLockRenewInterval is how often a holder extends its own lock. It has to divide
// FileLockMaxAge with room to spare, or a stalled write turns into a released lock.
const fileLockRenewInterval = time.Minute

// FileLockState is what a lock file records about its holder.
type FileLockState struct {
	Action    string    `json:"action"`
	PID       int       `json:"pid"`
	Host      string    `json:"host"`
	UpdatedAt time.Time `json:"updatedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Expired reports whether the lock has passed its own expiry, which is what makes a crashed
// holder release it. Time is read from this host, so the file is only meaningful to processes
// that share a clock - which is the case, since they share the storage path.
func (s FileLockState) Expired() bool {
	return !s.ExpiresAt.After(time.Now())
}

// String describes the holder for a message telling an operator what is in the way.
func (s FileLockState) String() string {
	return fmt.Sprintf("%s, running as process %d on %s since %s",
		s.Action, s.PID, s.Host, s.UpdatedAt.UTC().Format(time.RFC3339))
}

// FileLock is an advisory lock that processes sharing a storage path can see, unlike the
// in-process worker activities: a CLI command and a server run in different processes and
// would otherwise write the same rows without either noticing.
type FileLock struct {
	fileName string
	action   string
	done     chan struct{}
	stop     sync.Once
}

// ReadFileLock returns what a lock file records, or a zero state when it does not exist or
// cannot be read. An unreadable lock is treated as absent rather than as held, so a corrupt
// file cannot block an instance permanently.
func ReadFileLock(fileName string) FileLockState {
	var state FileLockState

	b, err := os.ReadFile(fileName) //nolint:gosec // path is derived from the storage path
	if err != nil {
		return state
	}

	if err = json.Unmarshal(b, &state); err != nil {
		return FileLockState{}
	}

	return state
}

// FileLockHeld returns a description of the process currently holding the lock, or "" when it
// is free. Callers use it to hold off on writes rather than to take the lock.
func FileLockHeld(fileName string) string {
	if fileName == "" {
		return ""
	}

	if state := ReadFileLock(fileName); !state.Expired() {
		return state.String()
	}

	return ""
}

// AcquireFileLock takes the named lock for the specified action and keeps renewing it until it
// is released. It reports an error naming the current holder when one is live.
func AcquireFileLock(fileName, action string) (*FileLock, error) {
	if fileName == "" {
		return nil, fmt.Errorf("lock filename is empty")
	}

	if held := FileLockHeld(fileName); held != "" {
		return nil, fmt.Errorf("%s is already in progress (%s)", action, held)
	}

	if err := fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return nil, err
	}

	lock := &FileLock{fileName: fileName, action: action, done: make(chan struct{})}

	if err := lock.write(); err != nil {
		return nil, err
	}

	go lock.renew()

	return lock, nil
}

// Release stops renewing the lock and removes its file. It is safe to call more than once, so
// a caller can defer it and still release early.
func (l *FileLock) Release() {
	if l == nil {
		return
	}

	l.stop.Do(func() {
		close(l.done)

		if err := os.Remove(l.fileName); err != nil && !os.IsNotExist(err) {
			log.Warnf("mutex: %s (release %s lock)", err, l.action)
		}
	})
}

// write records the holder and moves the expiry forward.
func (l *FileLock) write() error {
	host, _ := os.Hostname()
	now := time.Now()

	b, err := json.Marshal(FileLockState{
		Action:    l.action,
		PID:       os.Getpid(),
		Host:      host,
		UpdatedAt: now,
		ExpiresAt: now.Add(FileLockMaxAge),
	})

	if err != nil {
		return err
	}

	return os.WriteFile(l.fileName, b, fs.ModeFile)
}

// renew extends the lock until it is released. A write that fails is reported and retried at
// the next tick: the run is already underway, and abandoning it over a transient write error
// would be worse than letting the lock lapse.
func (l *FileLock) renew() {
	ticker := time.NewTicker(fileLockRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.done:
			return
		case <-ticker.C:
			if err := l.write(); err != nil {
				log.Warnf("mutex: %s (renew %s lock)", err, l.action)
			}
		}
	}
}
