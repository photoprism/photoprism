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
		"name":    txt.Clip(name, txt.ClipUserName),
		"browser": txt.Clip(browser, txt.ClipLog),
		"message": txt.Clip(message, txt.ClipLog),
	}
}

// LoginSuccess publishes a successful login event.
func LoginSuccess(ip, realm, name, browser string) {
	Publish("audit.login", LoginData(logrus.InfoLevel, ip, realm, name, browser, ""))
}

// LoginError publishes a login error event.
func LoginError(ip, realm, name, browser, error string) {
	Publish("audit.login", LoginData(logrus.ErrorLevel, ip, realm, name, browser, error))
}
