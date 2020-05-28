/*
This package encapsulates the geo location APIs.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package maps

import (
	"github.com/photoprism/photoprism/internal/event"
)

//go:generate go run gen.go
//go:generate go fmt .

var log = event.Log
