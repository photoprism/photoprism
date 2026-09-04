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

	// modTime is when the file was last written, read from the filesystem rather than from the
	// holder, so an expiry a skewed clock produced can be bounded by something local.
	modTime time.Time
}

// Expired reports whether the lock has passed its own expiry, which is what makes a crashed
// holder release it.
//
// The expiry is bounded by one interval past the file's modification time, because both timestamps
// inside it come from the holder's clock: one running fast would otherwise wedge every worker for
// as long as it was wrong. The kernel wrote the modification time, so a stale lock can be trusted.
func (s FileLockState) Expired() bool {
	expires := s.ExpiresAt

	switch {
	case s.modTime.IsZero():
		// No file, so there is nothing to age.
	case expires.IsZero():
		// The file exists but records no expiry, which is what a holder that has created it and
		// not yet written its state looks like. Reading that as free hands the lock to a second
		// caller in exactly the window creating it is meant to close, so its age decides instead.
		expires = s.modTime.Add(FileLockMaxAge)
	default:
		if limit := s.modTime.Add(FileLockMaxAge); expires.After(limit) {
			expires = limit
		}
	}

	return !expires.After(time.Now())
}

// HeldBy reports whether the state records this process on this host, so that a holder whose
// own lock lapsed and was taken over cannot release the one that replaced it.
func (s FileLockState) HeldBy(pid int, host string) bool {
	return s.PID == pid && s.Host == host
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
		state = FileLockState{}
	}

	// Stat even when the contents could not be read, because a lock file that exists is evidence
	// in itself: its age is what tells an unparsable one apart from an abandoned one.
	if info, statErr := os.Stat(fileName); statErr == nil {
		state.modTime = info.ModTime()
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
//
// The file is created exclusively, so two processes reaching this at the same moment cannot both
// come away holding it - which a read followed by a write allowed, and which is the one thing
// this lock exists to prevent.
func AcquireFileLock(fileName, action string) (*FileLock, error) {
	if fileName == "" {
		return nil, fmt.Errorf("lock filename is empty")
	}

	if err := fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return nil, err
	}

	lock := &FileLock{fileName: fileName, action: action, done: make(chan struct{})}

	if err := lock.create(); err != nil {
		return nil, err
	}

	go lock.renew()

	return lock, nil
}

// create takes the lock exclusively, replacing a lock whose holder let it expire.
func (l *FileLock) create() error {
	b, err := l.state()

	if err != nil {
		return err
	}

	f, err := os.OpenFile(l.fileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fs.ModeFile) //nolint:gosec // path derived from the storage path

	switch {
	case err == nil:
		defer f.Close()

		if _, err = f.Write(b); err != nil {
			return err
		}

		return nil
	case !os.IsExist(err):
		return err
	}

	// Somebody holds it. Only an expired lock may be taken over, and the replacement is renamed
	// into place so a reader never sees the gap between truncating and writing.
	if held := FileLockHeld(l.fileName); held != "" {
		return fmt.Errorf("%s is already in progress (%s)", l.action, held)
	}

	if err = l.replace(b); err != nil {
		return err
	}

	// Two processes taking over the same expired lock both rename, and the last one wins the
	// file. Reading it back is what tells the loser it does not hold what it just wrote.
	host, _ := os.Hostname()

	if state := ReadFileLock(l.fileName); !state.HeldBy(os.Getpid(), host) {
		return fmt.Errorf("%s is already in progress (%s)", l.action, state.String())
	}

	return nil
}

// Release stops renewing the lock and removes its file. It is safe to call more than once, so
// a caller can defer it and still release early.
//
// A lock this process no longer holds is left alone: one whose renewals failed may have been
// taken over legitimately, and removing it would free a run that is still working.
func (l *FileLock) Release() {
	if l == nil {
		return
	}

	l.stop.Do(func() {
		close(l.done)

		host, _ := os.Hostname()

		if state := ReadFileLock(l.fileName); !state.HeldBy(os.Getpid(), host) {
			return
		}

		if err := os.Remove(l.fileName); err != nil && !os.IsNotExist(err) {
			log.Warnf("mutex: %s (release %s lock)", err, l.action)
		}
	})
}

// state returns this holder's lock file contents, with the expiry moved forward.
func (l *FileLock) state() ([]byte, error) {
	host, _ := os.Hostname()
	now := time.Now()

	return json.Marshal(FileLockState{
		Action:    l.action,
		PID:       os.Getpid(),
		Host:      host,
		UpdatedAt: now,
		ExpiresAt: now.Add(FileLockMaxAge),
	})
}

// write records the holder and moves the expiry forward.
func (l *FileLock) write() error {
	b, err := l.state()

	if err != nil {
		return err
	}

	return l.replace(b)
}

// replace writes the lock file through a temporary file in the same directory, so that a reader
// sees either the previous contents or the new ones.
//
// Writing in place leaves the file empty between truncating and writing, and an unreadable lock
// reads as free - which over an hour of renewals is a window every worker polls into.
func (l *FileLock) replace(b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(l.fileName), filepath.Base(l.fileName)+".*")

	if err != nil {
		return err
	}

	tempName := f.Name()

	defer func() {
		if removeErr := os.Remove(tempName); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warnf("mutex: %s (remove %s lock temp file)", removeErr, l.action)
		}
	}()

	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Chmod(tempName, fs.ModeFile); err != nil {
		return err
	}

	return os.Rename(tempName, l.fileName)
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
