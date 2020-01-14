/*
This package encapsulates image metadata decoding and conversion to/from XMP and Exif.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package meta

import (
	"github.com/dsoprea/go-exif/v2"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var im *exif.IfdMapping

func init() {
	im = exif.NewIfdMapping()

	if err := exif.LoadStandardIfds(im); err != nil {
		log.Errorf("meta: %s", err.Error())
	}
}
