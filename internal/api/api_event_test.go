package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
)

func TestEventString(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "updated", StatusUpdated.String())
		assert.Equal(t, "created", StatusCreated.String())
	})
}

func TestPublishEntityEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sub := event.Subscribe("photos.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		publishEntityEvent("photos", StatusUpdated, "pqkm36fjqvset9uy")

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "photos.updated", msg.Name)
			uids, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{"pqkm36fjqvset9uy"}, uids)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one photos.updated event")
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		sub := event.Subscribe("photos.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		publishEntityEvent("photos", StatusUpdated, "not a uid!")

		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event for invalid uid: %s %v", msg.Name, msg.Fields)
		case <-time.After(200 * time.Millisecond):
			// expected: nothing published.
		}
	})
}

func TestPublishPhotoEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sub := event.Subscribe("photos.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishPhotoEvent(StatusUpdated, "pqkm36fjqvset9uy")

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "photos.updated", msg.Name)
			uids, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{"pqkm36fjqvset9uy"}, uids)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one photos.updated event")
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		sub := event.Subscribe("photos.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishPhotoEvent(StatusUpdated, "")

		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event for empty uid: %s %v", msg.Name, msg.Fields)
		case <-time.After(200 * time.Millisecond):
			// expected: nothing published.
		}
	})
}

func TestPublishAlbumEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sub := event.Subscribe("albums.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishAlbumEvent(StatusUpdated, "as6sg6bxpogaaba7")

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "albums.updated", msg.Name)
			uids, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{"as6sg6bxpogaaba7"}, uids)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one albums.updated event")
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		sub := event.Subscribe("albums.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishAlbumEvent(StatusUpdated, "not a uid!")

		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event for invalid uid: %s %v", msg.Name, msg.Fields)
		case <-time.After(200 * time.Millisecond):
			// expected: nothing published.
		}
	})
}

func TestPublishLabelEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sub := event.Subscribe("labels.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishLabelEvent(StatusUpdated, "ls6sg6b1wowuy3c2")

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "labels.updated", msg.Name)
			uids, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{"ls6sg6b1wowuy3c2"}, uids)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one labels.updated event")
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		sub := event.Subscribe("labels.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishLabelEvent(StatusUpdated, "")

		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event for empty uid: %s %v", msg.Name, msg.Fields)
		case <-time.After(200 * time.Millisecond):
			// expected: nothing published.
		}
	})
}

func TestPublishSubjectEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sub := event.Subscribe("subjects.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishSubjectEvent(StatusUpdated, "js6sg6b1qekk9jx8")

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "subjects.updated", msg.Name)
			uids, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{"js6sg6b1qekk9jx8"}, uids)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one subjects.updated event")
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		sub := event.Subscribe("subjects.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		PublishSubjectEvent(StatusUpdated, "not a uid!")

		select {
		case msg := <-sub.Receiver:
			t.Fatalf("unexpected event for invalid uid: %s %v", msg.Name, msg.Fields)
		case <-time.After(200 * time.Millisecond):
			// expected: nothing published.
		}
	})
}
