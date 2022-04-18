package config

import (
	"errors"
)

var (
	ErrReadOnly = errors.New("not available in read-only mode")
)

func LogError(err error) {
	if err != nil {
		log.Errorf("config: %s", err.Error())
	}
}
