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

// Error publishes an error notification with the given message.
func Error(msg string) {
	Log.Error(strings.ToLower(msg))
	Publish("notify.error", Data{"message": msg})
}

// Success publishes a success notification with the given message.
func Success(msg string) {
	Log.Info(strings.ToLower(msg))
	Publish("notify.success", Data{"message": msg})
}

// Info publishes an informational notification with the given message.
func Info(msg string) {
	Log.Info(strings.ToLower(msg))
	Publish("notify.info", Data{"message": msg})
}

// Warn publishes a warning notification with the given message.
func Warn(msg string) {
	Log.Warn(strings.ToLower(msg))
	Publish("notify.warning", Data{"message": msg})
}

// ErrorMsg publishes a localized error notification.
func ErrorMsg(id i18n.Message, params ...interface{}) {
	Error(i18n.Msg(id, params...))
}

// SuccessMsg publishes a localized success notification.
func SuccessMsg(id i18n.Message, params ...interface{}) {
	Success(i18n.Msg(id, params...))
}

// InfoMsg publishes a localized informational notification.
func InfoMsg(id i18n.Message, params ...interface{}) {
	Info(i18n.Msg(id, params...))
}

// WarnMsg publishes a localized warning notification.
func WarnMsg(id i18n.Message, params ...interface{}) {
	Warn(i18n.Msg(id, params...))
}
