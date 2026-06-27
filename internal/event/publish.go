package event

import (
	"strings"

	"github.com/sirupsen/logrus"

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

// publishMsg logs and publishes a localized notification, carrying the rendered message plus the
// untranslated source id and params so the frontend can render it in the user's current UI language.
func publishMsg(level logrus.Level, topic string, id i18n.Message, params ...any) {
	msg := i18n.Msg(id, params...)
	Log.Log(level, strings.ToLower(msg))
	Publish(topic, Data{"message": msg, "messageId": i18n.Source(id), "messageParams": params})
}

// ErrorMsg publishes a localized error notification.
func ErrorMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.ErrorLevel, "notify.error", id, params...)
}

// SuccessMsg publishes a localized success notification.
func SuccessMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.InfoLevel, "notify.success", id, params...)
}

// InfoMsg publishes a localized informational notification.
func InfoMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.InfoLevel, "notify.info", id, params...)
}

// WarnMsg publishes a localized warning notification.
func WarnMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.WarnLevel, "notify.warning", id, params...)
}
