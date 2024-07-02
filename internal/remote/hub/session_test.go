package hub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_Expired(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		session := Session{
			MapKey:    "",
			ExpiresAt: "",
		}
		assert.True(t, session.Expired())
	})
	t.Run("true", func(t *testing.T) {
		date := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
		session := Session{
			MapKey:    "",
			ExpiresAt: date.String(),
		}
		assert.True(t, session.Expired())
	})
}
