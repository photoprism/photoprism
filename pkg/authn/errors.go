package authn

import (
	"errors"
)

var (
	ErrPasscodeRequired         = errors.New("passcode required")
	ErrPasscodeNotSetUp         = errors.New("passcode required, but not set up")
	ErrPasscodeNotVerified      = errors.New("passcode not verified")
	ErrPasscodeAlreadyActivated = errors.New("passcode already activated")
	ErrPasscodeNotSupported     = errors.New("passcode not supported")
	ErrInvalidPasscode          = errors.New("invalid passcode")
	ErrInvalidPasscodeFormat    = errors.New("invalid passcode format")
	ErrInvalidPasscodeKey       = errors.New("invalid passcode key")
	ErrInvalidPasscodeType      = errors.New("invalid passcode type")
	ErrPasswordRequired         = errors.New("password required")
	ErrInvalidPassword          = errors.New("invalid password")
	ErrInvalidPasswordFormat    = errors.New("invalid password format")
	ErrInvalidCredentials       = errors.New("invalid credentials")
)
