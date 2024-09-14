package event

import (
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/pkg/txt"
)

// LoginData returns a login event message.
func LoginData(level logrus.Level, ip, realm, name, browser, message string) Data {
	return Data{
		"time":    TimeStamp(),
		"level":   level.String(),
		"ip":      txt.Clip(ip, txt.ClipIP),
		"realm":   txt.Clip(realm, txt.ClipRealm),
		"name":    txt.Clip(name, txt.ClipUsername),
		"browser": txt.Clip(browser, txt.ClipLog),
		"message": txt.Clip(message, txt.ClipLog),
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
