/*
Package classify encapsulates image classification functionnality using TensorFlow

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package classify

import (
	"github.com/photoprism/photoprism/internal/event"
)

//go:generate go run gen.go
//go:generate go fmt .

var log = event.Log
