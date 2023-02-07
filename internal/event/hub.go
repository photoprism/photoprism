package event

import (
	"strings"

	"github.com/leandro-lugaresi/hub"
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
