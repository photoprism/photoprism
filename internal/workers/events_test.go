package workers

import (
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
)

// subscribePhotoEvents subscribes to photos.<action> events for the duration of a test.
func subscribePhotoEvents(t *testing.T, action string) hub.Subscription {
	t.Helper()

	sub := event.Subscribe("photos." + action)
	t.Cleanup(func() { event.Unsubscribe(sub) })

	return sub
}

// receivePhotoEvents drains the subscription and returns the UID list of each event received.
func receivePhotoEvents(t *testing.T, sub hub.Subscription) (events [][]string) {
	t.Helper()

	for {
		select {
		case msg := <-sub.Receiver:
			uids, ok := msg.Fields["entities"].([]string)
			require.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			events = append(events, uids)
		case <-time.After(250 * time.Millisecond):
			return events
		}
	}
}

// eventsAtLog is a log hook that records how many events a subscription held when a matching message was logged.
type eventsAtLog struct {
	sub   hub.Subscription
	match func(string) bool
	seen  bool
	count int
}

// Levels returns the log levels the hook receives.
func (h *eventsAtLog) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire records the number of buffered events for the first matching message.
func (h *eventsAtLog) Fire(entry *logrus.Entry) error {
	if !h.seen && h.match(entry.Message) {
		h.seen = true
		h.count = len(h.sub.Receiver)
	}

	return nil
}

// captureEventsAtLog replaces the package logger with one that counts the events buffered when match first holds.
func captureEventsAtLog(t *testing.T, sub hub.Subscription, match func(string) bool) *eventsAtLog {
	t.Helper()

	h := &eventsAtLog{sub: sub, match: match}
	logger, _ := test.NewNullLogger()
	logger.AddHook(h)

	prev := log
	log = logger
	t.Cleanup(func() { log = prev })

	return h
}
