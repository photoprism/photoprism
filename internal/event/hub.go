package event

import (
	"github.com/leandro-lugaresi/hub"
)

type Hub = hub.Hub
type Data = hub.Fields
type Message = hub.Message

var channelCap = 10
var sharedHub = NewHub()


func NewHub () *Hub {
	return hub.New()
}

func SharedHub() *Hub {
	return sharedHub
}

func Error(msg string) {
	Log.Error(msg)
	Publish("notify.error", Data{"msg": msg})
}

func Success(msg string) {
	Log.Info(msg)
	Publish("notify.success", Data{"msg": msg})
}

func Info(msg string) {
	Log.Info(msg)
	Publish("notify.info", Data{"msg": msg})
}

func Warning(msg string) {
	Log.Warn(msg)
	Publish("notify.warning", Data{"msg": msg})
}

func Publish (event string, data Data) {
	SharedHub().Publish(Message{
		Name:   event,
		Fields: data,
	})
}

func Subscribe(topics ...string) hub.Subscription {
	return SharedHub().Subscribe(channelCap, topics...)
}

func Unsubscribe(s hub.Subscription) {
	SharedHub().Unsubscribe(s)
}
