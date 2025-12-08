package fs

import (
	"errors"
	"io"
)

// Generic errors that may occur when accessing files and folders:
var (
	EOF                 = io.EOF
	ErrUnexpectedEOF    = io.ErrUnexpectedEOF
	ErrShortWrite       = io.ErrShortWrite
	ErrShortBuffer      = io.ErrShortBuffer
	ErrNoProgress       = io.ErrNoProgress
	ErrInvalidWrite     = errors.New("invalid write result")
	ErrPermissionDenied = errors.New("permission denied")
)
