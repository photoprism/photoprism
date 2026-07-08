package photoprism

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/pkg/log/status"
)

// walkResultLog returns the level, message, and emit flag for a directory-walk result.
// status.ErrCanceled surfaces at info level (user-initiated stop); status.ErrInsufficientStorage is
// suppressed because the storage-check helper already logged the actionable cause.
func walkResultLog(prefix string, err error) (level logrus.Level, message string, emit bool) {
	switch {
	case err == nil:
		return 0, "", false
	case errors.Is(err, status.ErrCanceled):
		return logrus.InfoLevel, fmt.Sprintf("%s: %s", prefix, status.Canceled), true
	case errors.Is(err, status.ErrInsufficientStorage):
		return 0, "", false
	default:
		return logrus.ErrorLevel, fmt.Sprintf("%s: %s", prefix, err.Error()), true
	}
}

// logWalkResult emits a log line for a directory-walk result via walkResultLog.
func logWalkResult(prefix string, err error) {
	if level, msg, emit := walkResultLog(prefix, err); emit {
		log.Log(level, msg)
	}
}
