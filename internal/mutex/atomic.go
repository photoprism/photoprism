package mutex

import (
	"sync/atomic"
)

// Restart signals that the application should be restarted,
// e.g. after an update or a config changes.
var Restart = atomic.Bool{}
