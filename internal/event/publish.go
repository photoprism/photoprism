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

// notifyMsg publishes a localized notification without logging it. The payload carries the
// rendered message plus the untranslated source id and params, so the frontend can render it
// in the user's current UI language.
func notifyMsg(topic string, id i18n.Message, params ...any) {
	Publish(topic, Data{"message": i18n.Msg(id, params...), "messageId": i18n.Source(id), "messageParams": params})
}

// publishMsg logs and publishes a localized notification.
// The log line stays English through i18n.Lower, so server logs do not follow the instance locale.
func publishMsg(level logrus.Level, topic string, id i18n.Message, params ...any) {
	Log.Log(level, i18n.Lower(id, params...))
	notifyMsg(topic, id, params...)
}

// ErrorMsg publishes a localized error notification.
func ErrorMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.ErrorLevel, "notify.error", id, params...)
}

// SuccessMsg publishes a localized success notification.
func SuccessMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.InfoLevel, "notify.success", id, params...)
}

// PublishSuccessMsg publishes a localized success notification without logging it,
// for callers that write a more specific log line themselves.
func PublishSuccessMsg(id i18n.Message, params ...any) {
	notifyMsg("notify.success", id, params...)
}

// InfoMsg publishes a localized informational notification.
func InfoMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.InfoLevel, "notify.info", id, params...)
}

// WarnMsg publishes a localized warning notification.
func WarnMsg(id i18n.Message, params ...any) {
	publishMsg(logrus.WarnLevel, "notify.warning", id, params...)
}

// PublishCompleted notifies subscribed clients that an index, import or upload run has finished.
// It carries no path: every consumer treats these as "something changed, refresh", and topic
// routing delivers them to every session allowed to subscribe rather than to the one that acted.
func PublishCompleted(topics []string, uid, action string, elapsed int) {
	data := Data{"uid": uid, "seconds": elapsed}

	if action != "" {
		data["action"] = action
	}

	for _, topic := range topics {
		Publish(topic, data)
	}
}
