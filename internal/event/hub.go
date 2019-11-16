package event

import (
	"github.com/leandro-lugaresi/hub"
	"github.com/sirupsen/logrus"
)

type Hub = hub.Hub
type Data = hub.Fields
type Message = hub.Message
var log *logrus.Logger
var channelCap = 10
var sharedHub = NewHub()

func init() {
	log = logrus.StandardLogger()
}

func NewHub () *Hub {
	return hub.New()
}

func SharedHub() *Hub {
	return sharedHub
}

func Error(msg string) {
	log.Error(msg)
	Publish("alert.error", Data{"msg": msg})
}

func Success(msg string) {
	log.Info(msg)
	Publish("alert.success", Data{"msg": msg})
}

func Info(msg string) {
	log.Info(msg)
	Publish("alert.info", Data{"msg": msg})
}

func Warning(msg string) {
	log.Warn(msg)
	Publish("alert.warning", Data{"msg": msg})
}

func Publish (event string, data Data) {
	log.Infof("publish %s: %v", event, data)
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
