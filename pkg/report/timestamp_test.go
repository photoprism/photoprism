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

	t.Run("1665389030", func(t *testing.T) {
		now := time.Unix(1665389030, 0)
		assert.Equal(t, "2022-10-10 08:03:50", DateTime(&now))
	})
}

func TestUnixTime(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", UnixTime(0))
	})

	t.Run("1665389030", func(t *testing.T) {
		assert.Equal(t, "2022-10-10 08:03:50", UnixTime(1665389030))
	})
}
