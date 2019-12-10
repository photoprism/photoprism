/*
Package config contains filesystem related utility functions.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package util

import (
	"github.com/photoprism/photoprism/internal/event"
)

//go:generate go run gen.go

var log = event.Log
