/*
Package entity contains models for data storage based on GORM.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var GeoApi = "places"

// logError logs the message if the argument is an error.
func logError(result *gorm.DB) {
	if result.Error != nil {
		log.Error(result.Error.Error())
	}
}
