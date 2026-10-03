package photoprism

import (
	"sync"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

var (
	// transcodeFallbacks remembers the hardware encoders that failed since they last succeeded.
	transcodeFallbacks   = make(map[encode.Encoder]struct{})
	transcodeFallbacksMu sync.Mutex
)

// firstTranscodeFallback records a failed hardware encoder and reports whether it is the first failure since
// the encoder last succeeded. A GPU that cannot be used fails for every file, so later failures are not warnings.
func firstTranscodeFallback(encoder encode.Encoder) bool {
	transcodeFallbacksMu.Lock()
	defer transcodeFallbacksMu.Unlock()

	if _, found := transcodeFallbacks[encoder]; found {
		return false
	}

	transcodeFallbacks[encoder] = struct{}{}

	return true
}

// clearTranscodeFallback forgets the recorded failure of an encoder after it succeeded.
func clearTranscodeFallback(encoder encode.Encoder) {
	transcodeFallbacksMu.Lock()
	defer transcodeFallbacksMu.Unlock()

	delete(transcodeFallbacks, encoder)
}

// resetTranscodeFallbacks forgets the logged fallbacks, so that tests start from the state of a new process.
func resetTranscodeFallbacks() {
	transcodeFallbacksMu.Lock()
	defer transcodeFallbacksMu.Unlock()

	clear(transcodeFallbacks)
}
