package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", DateTime(nil))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", DateTime(&time.Time{}))
	})
	t.Run("Num1665389030", func(t *testing.T) {
		now := time.Unix(1665389030, 0)
		assert.Equal(t, "2022-10-10 08:03:50", DateTime(&now))
	})
}

func TestUnixTime(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", UnixTime(0))
	})
	t.Run("Num1665389030", func(t *testing.T) {
		assert.Equal(t, "2022-10-10 08:03:50", UnixTime(1665389030))
	})
}

func TestDate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		born := time.Date(1981, 1, 22, 14, 30, 0, 0, time.UTC)
		// The time of day is dropped rather than rendered, since a birth date does not carry one.
		assert.Equal(t, "1981-01-22", Date(&born))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", Date(nil))
	})
	t.Run("Zero", func(t *testing.T) {
		// How "not set" reaches this, so it must not render as year one.
		var zero time.Time
		assert.Equal(t, "", Date(&zero))
	})
}
