//go:build debug
// +build debug

package hub

import (
	"os"

	"github.com/photoprism/photoprism/pkg/clean"
)

func init() {
	if debugUrl := os.Getenv("PHOTOPRISM_HUB_URL"); debugUrl != "" {
		log.Infof("config: set hub url to %s", clean.Log(debugUrl))
		ServiceURL = debugUrl
	}
}
