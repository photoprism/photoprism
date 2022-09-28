package event

import (
	"strings"

	"github.com/leandro-lugaresi/hub"

	"github.com/photoprism/photoprism/internal/i18n"
)

type Hub = hub.Hub
type Data = hub.Fields
type Message = hub.Message

const TopicSep = "."

var channelCap = 100
var sharedHub = NewHub()

func NewHub() *Hub {
	return hub.New()
}

func SharedHub() *Hub {
	return sharedHub
}

// Publish publishes a message to all subscribers.
func Publish(event string, data Data) {
	SharedHub().Publish(Message{
		Name:   event,
		Fields: data,
	})
}

// Subscribe creates a topic subscription and returns i
func Subscribe(topics ...string) hub.Subscription {
	return SharedHub().NonBlockingSubscribe(channelCap, topics...)
}

// Unsubscribe deletes the subscription of a topic.
func Unsubscribe(s hub.Subscription) {
	SharedHub().Unsubscribe(s)
}

// Topic splits the topic name into the channel and event names.
func Topic(topic string) (ch, ev string) {
	ch, ev, _ = strings.Cut(topic, TopicSep)
	return ch, ev
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
