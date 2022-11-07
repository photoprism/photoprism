package txt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeStamp(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", TimeStamp(nil))
	})

	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", TimeStamp(&time.Time{}))
	})

	t.Run("1665389030", func(t *testing.T) {
		now := time.Unix(1665389030, 0)
		assert.Equal(t, "2022-10-10 08:03:50", TimeStamp(&now))
	})
}

func TestNTimes(t *testing.T) {
	t.Run("-2", func(t *testing.T) {
		assert.Equal(t, "", NTimes(-2))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, "", NTimes(-1))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, "", NTimes(0))
	})
	t.Run("1", func(t *testing.T) {
		assert.Equal(t, "", NTimes(1))
	})
	t.Run("999", func(t *testing.T) {
		assert.Equal(t, "999 times", NTimes(999))
	})
}
