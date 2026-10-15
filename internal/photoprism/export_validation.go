package photoprism

import (
	"sync"
	"time"
)

// exportValidation retains successful hashes and shares in-flight verification of one file version.
var exportValidation = exportValidationCache{entries: make(map[string]*exportValidationCall)}

// exportValidationSlots bounds full source reads independently of lightweight container probes.
var exportValidationSlots = make(chan struct{}, 2)

// exportValidationCall carries the result of one source verification.
type exportValidationCall struct {
	done    chan struct{}
	err     error
	expires time.Time
}

// exportValidationCache bounds both retained results and in-flight source verifications.
type exportValidationCache struct {
	mutex   sync.Mutex
	entries map[string]*exportValidationCall
}

// verify reuses successful version checks without caching failures or holding a lock during I/O.
func (cache *exportValidationCache) verify(key string, run func() error) (err error) {
	cache.mutex.Lock()
	if call := cache.entries[key]; call != nil && (call.expires.IsZero() || time.Now().Before(call.expires)) {
		cache.mutex.Unlock()
		<-call.done
		return call.err
	}
	delete(cache.entries, key)
	if len(cache.entries) >= 128 {
		for name, call := range cache.entries {
			if !call.expires.IsZero() {
				delete(cache.entries, name)
				break
			}
		}
	}
	select {
	case exportValidationSlots <- struct{}{}:
	default:
		cache.mutex.Unlock()
		return ErrExportBusy
	}
	call := &exportValidationCall{done: make(chan struct{})}
	if cache.entries == nil {
		cache.entries = make(map[string]*exportValidationCall)
	}
	cache.entries[key] = call
	cache.mutex.Unlock()
	completed := false
	defer func() {
		cache.mutex.Lock()
		if !completed {
			err = ErrExportSourceChanged
		}
		call.err = err
		if err == nil {
			call.expires = time.Now().Add(5 * time.Minute)
		} else {
			delete(cache.entries, key)
		}
		close(call.done)
		cache.mutex.Unlock()
		<-exportValidationSlots
	}()
	err = run()
	completed = true
	return err
}
