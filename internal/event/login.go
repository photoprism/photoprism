package event

import (
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/pkg/txt"
)

// LoginData returns a login event message.
func LoginData(level logrus.Level, ip, username, useragent, error string) Data {
	return Data{
		"time":      TimeStamp(),
		"level":     level.String(),
		"ip":        txt.Clip(ip, txt.ClipIP),
		"message":   txt.Clip(error, txt.ClipLog),
		"username":  txt.Clip(username, txt.ClipUserName),
		"useragent": txt.Clip(useragent, txt.ClipLog),
	}
}

// LoginSuccess publishes a successful login event.
func LoginSuccess(ip, username, useragent string) {
	Publish("audit.login", LoginData(logrus.InfoLevel, ip, username, useragent, ""))
}

// LoginError publishes a login error event.
func LoginError(ip, username, useragent, error string) {
	Publish("audit.login", LoginData(logrus.ErrorLevel, ip, username, useragent, error))
}
