package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLang(t *testing.T) {
	assert.Equal(t, English, Lang)
	SetLang("D")
	assert.Equal(t, English, Lang)
	SetLang("De")
	assert.Equal(t, German, Lang)
	SetLang("")
	assert.Equal(t, English, Lang)
	assert.Equal(t, Default, Lang)
}
