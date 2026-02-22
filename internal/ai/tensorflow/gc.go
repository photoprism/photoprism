package tensorflow

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
)

const gcEveryDefault uint64 = 200

var (
	gcEvery   = gcEveryDefault
	gcCounter uint64
)

func init() {
	if v := strings.TrimSpace(os.Getenv("PHOTOPRISM_TF_GC_EVERY")); v != "" {
		if strings.HasPrefix(v, "-") {
			gcEvery = 0
			return
		}

		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			gcEvery = n
		}
	}
}

// MaybeCollectTensorMemory triggers GC and returns freed C-allocated tensor memory
// to the OS every gcEvery calls; set gcEvery to 0 to disable the throttling.
func MaybeCollectTensorMemory() {
	if gcEvery == 0 {
		return
	}

	if atomic.AddUint64(&gcCounter, 1)%gcEvery != 0 {
		return
	}

	debug.FreeOSMemory()
}
