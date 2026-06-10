package clean

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeSet(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, TimeSet(nil))
	})
	t.Run("Zero", func(t *testing.T) {
		zero := time.Time{}
		assert.False(t, TimeSet(&zero))
	})
	t.Run("Set", func(t *testing.T) {
		now := time.Now()
		assert.True(t, TimeSet(&now))
	})
}
