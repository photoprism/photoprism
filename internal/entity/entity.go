/*
This package contains models for data storage based on GORM.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/util"
)

var log = event.Log

func logError(result *gorm.DB) {
	if result.Error != nil {
		log.Error(result.Error.Error())
	}
}

func ID(prefix rune) string {
	result := make([]byte, 0, 17)
	result = append(result, byte(prefix))
	result = append(result, strconv.FormatInt(time.Now().UTC().Unix(), 36)[0:6]...)
	result = append(result, util.RandomToken(10)...)

	return string(result)
}
