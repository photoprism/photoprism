package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseName(t *testing.T) {
	t.Run("BillGates", func(t *testing.T) {
		result := ParseName("William Henry Gates III")
		t.Logf("Name: %#v", result)
		assert.Equal(t, "William", result.Given)
	})
}
