package event

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/i18n"
)

// Publish publishes a message to all subscribers.
func Publish(event string, data Data) {
	SharedHub().Publish(Message{
		Name:   event,
		Fields: data,
	})
}

func Error(msg string) {
	Log.Error(strings.ToLower(msg))
	Publish("notify.error", Data{"message": msg})
}

func Success(msg string) {
	Log.Info(strings.ToLower(msg))
	Publish("notify.success", Data{"message": msg})
}

func Info(msg string) {
	Log.Info(strings.ToLower(msg))
	Publish("notify.info", Data{"message": msg})
}

func Warn(msg string) {
	Log.Warn(strings.ToLower(msg))
	Publish("notify.warning", Data{"message": msg})
}

func ErrorMsg(id i18n.Message, params ...interface{}) {
	Error(i18n.Msg(id, params...))
}

func SuccessMsg(id i18n.Message, params ...interface{}) {
	Success(i18n.Msg(id, params...))
}

func InfoMsg(id i18n.Message, params ...interface{}) {
	Info(i18n.Msg(id, params...))
}

func WarnMsg(id i18n.Message, params ...interface{}) {
	Warn(i18n.Msg(id, params...))
}
