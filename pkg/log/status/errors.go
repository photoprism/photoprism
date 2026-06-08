package status

import "errors"

// ErrCanceled signals that a worker or walk callback was interrupted by Cancel().
var ErrCanceled = errors.New(Canceled)

// ErrInsufficientStorage signals that an operation cannot proceed because the
// configured quota is exhausted or the storage path is critically low on free disk space.
var ErrInsufficientStorage = errors.New(InsufficientStorage)
