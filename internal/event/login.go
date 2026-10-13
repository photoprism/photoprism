package event

import (
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// LoginData returns a login event message, with each value passed through clean.Display.
func LoginData(level logrus.Level, ip, realm, name, browser, message string) Data {
	return Data{
		"time":    TimeStamp(),
		"level":   level.String(),
		"ip":      clean.DisplayRunes(ip, txt.ClipIP),
		"realm":   clean.DisplayRunes(realm, txt.ClipRealm),
		"name":    clean.DisplayRunes(name, txt.ClipUsername),
		"browser": clean.DisplayRunes(browser, txt.ClipLog),
		"message": clean.DisplayRunes(message, txt.ClipLog),
	}
}

// LoginInfo publishes a successful login event.
func LoginInfo(ip, realm, name, browser string) {
	Publish("login.info", LoginData(logrus.InfoLevel, ip, realm, name, browser, ""))
}

// LoginError publishes a login error event.
func LoginError(ip, realm, name, browser, error string) {
	Publish("login.error", LoginData(logrus.ErrorLevel, ip, realm, name, browser, error))
}
