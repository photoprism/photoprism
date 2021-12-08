/*
Package entity provides models for storing index information based on the GORM library.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var GeoApi = "places"

// Log logs the error if any and keeps quiet otherwise.
func Log(model, action string, err error) {
	if err != nil {
		log.Errorf("%s: %s (%s)", model, err, action)
	}
}
