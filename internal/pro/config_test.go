package pro

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_MapKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewConfig("develop", "testdata/new.yml")
		assert.Equal(t, "", c.MapKey())
	})
}
