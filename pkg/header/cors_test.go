package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowCORS(t *testing.T) {
	t.Run("CSS", func(t *testing.T) {
		assert.False(t, AllowCORS(""))
		assert.False(t, AllowCORS("."))
		assert.False(t, AllowCORS(" "))
		assert.False(t, AllowCORS(".css"))
		assert.False(t, AllowCORS(" .css"))
		assert.True(t, AllowCORS("a.css"))
		assert.True(t, AllowCORS("static/files/styles.css"))
		assert.True(t, AllowCORS("/static/files/styles.css"))
		assert.True(t, AllowCORS("/static/files/a.css"))
		assert.False(t, AllowCORS("/static/files/styles/.css"))
		assert.False(t, AllowCORS("/.css"))
		assert.False(t, AllowCORS(".css"))
		assert.False(t, AllowCORS("css"))
	})
}
