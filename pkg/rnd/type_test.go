package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUidType(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		result, prefix := IdType("")
		assert.Equal(t, TypeEmpty, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("Unknown", func(t *testing.T) {
		result, prefix := IdType("lt9kr7ur57iru5i7uy3c2")
		assert.Equal(t, TypeUnknown, result)
		assert.Equal(t, PrefixNone, prefix)
	})
	t.Run("LabelUID", func(t *testing.T) {
		result, prefix := IdType("ls6sg1e1wowuy3c2")
		assert.Equal(t, TypeUID, result)
		assert.Equal(t, byte('l'), prefix)
	})
}
