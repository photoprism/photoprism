package rnd

import (
	uuid "github.com/satori/go.uuid"
)

// UUID returns a standard, random UUID as string.
func UUID() string {
	if id, err := uuid.NewV4(); err != nil {
		return ""
	} else {
		return id.String()
	}
}
