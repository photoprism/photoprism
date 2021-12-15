package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := Query("table spoon & usa | img% json OR BILL!")
		assert.Equal(t, "table spoon & usa | img* json|bill", q)
	})
}
