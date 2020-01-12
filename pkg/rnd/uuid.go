package rnd

import (
	uuid "github.com/satori/go.uuid"
)

// UUID returns a standard, random UUID as string.
func UUID() string {
	return uuid.NewV4().String()
}
