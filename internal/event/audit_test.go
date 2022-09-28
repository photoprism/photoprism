package event

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/clean"
)

func TestFormat(t *testing.T) {
	assert.Equal(t, "", Format(nil))
	assert.Equal(t, "user", Format([]string{"user"}))
	assert.Equal(t, "user › michael › not found", Format([]string{"user", "michael", "not found"}))

	result := Format([]string{"user", "%s not found"}, clean.LogQuote("michael"))
	expected := "user › 'michael' not found"

	assert.Equal(t, expected, result)

	t.Log(result)
}
