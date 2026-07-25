package authtoken

import "errors"

// Errors returned when a token cannot be verified.
var (
	ErrMalformed = errors.New("authtoken: malformed token")
	ErrSignature = errors.New("authtoken: invalid signature")
	ErrExpired   = errors.New("authtoken: token expired")
)
