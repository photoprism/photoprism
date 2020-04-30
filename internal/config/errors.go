package config

import (
	"errors"
)

// Define photoprism specific errors
var (
	ErrReadOnly     = errors.New("not available in read-only mode")
	ErrUnauthorized = errors.New("please log in and try again")
	ErrUploadNSFW   = errors.New("upload might be offensive")
)

func LogError(err error) {
	if err != nil {
		log.Errorf("config: %s", err.Error())
	}
}
