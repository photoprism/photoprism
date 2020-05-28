/*
This package contains the PhotoPrism REST api.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package api

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

func logError(prefix string, err error) {
	if err != nil {
		log.Errorf("%s: %s", prefix, err.Error())
	}
}

func UpdateClientConfig(conf *config.Config) {
	event.Publish("config.updated", event.Data{"config": conf.ClientConfig()})
}
