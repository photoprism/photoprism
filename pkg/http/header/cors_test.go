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
	t.Run("Scripts", func(t *testing.T) {
		assert.True(t, AllowCORS("/static/build/app.c95b33ae.js"))
		// The pdf.js worker is emitted as .mjs, and a CDN-hosted bundle can only load it
		// through a cross-origin import, which needs the CORS header.
		assert.True(t, AllowCORS("/static/build/6d70933f83589315b6d9.mjs"))
		assert.False(t, AllowCORS("/static/build/.mjs"))
		assert.False(t, AllowCORS("mjs"))
	})
	t.Run("NotAllowed", func(t *testing.T) {
		assert.False(t, AllowCORS("/static/build/sw.map"))
		assert.False(t, AllowCORS("/originals/photo.jpg"))
	})
}
