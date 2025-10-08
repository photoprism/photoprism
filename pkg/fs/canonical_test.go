package fs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNonCanonical(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, true, NonCanonical("short"))
	})
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, true, NonCanonical("short/short/short/test1_"))
	})
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, true, NonCanonical("short#short_short_test1?"))
	})
	t.Run("ShortTestTestTest1234", func(t *testing.T) {
		assert.Equal(t, false, NonCanonical("hort/test_test_test12345"))
	})
}

func TestCanonicalName(t *testing.T) {
	date := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

	assert.Equal(t, "20091117_203458_EEEEEEEE", CanonicalName(date, "123", ""))
	assert.Equal(t, "20091117_203458_12345678", CanonicalName(date, "12345678", ""))
	assert.Equal(t, "091117-203458_12345678", CanonicalName(date, "12345678", "060102-150405_"))
}
