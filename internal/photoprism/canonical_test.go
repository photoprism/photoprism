package photoprism

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNonCanonical(t *testing.T) {
	t.Run("short", func(t *testing.T) {
		assert.Equal(t, true, NonCanonical("short"))
	})
	t.Run("short_", func(t *testing.T) {
		assert.Equal(t, true, NonCanonical("short_"))
	})
	t.Run("short?", func(t *testing.T) {
		assert.Equal(t, true, NonCanonical("short?"))
	})
	t.Run("short/test_test_test1234", func(t *testing.T) {
		assert.Equal(t, false, NonCanonical("hort/test_test_test12345"))
	})
}

func TestCanonicalName(t *testing.T) {
	date := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

	assert.Equal(t, "20091117_203458_ERROR000", CanonicalName(date, "123"))
	assert.Equal(t, "20091117_203458_12345678", CanonicalName(date, "12345678"))
}
