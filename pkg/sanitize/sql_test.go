package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlString(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlString(""))
	})
	t.Run("Special", func(t *testing.T) {
		s := "' \" \t \n %_''"
		exp := "\\' \\\"   %\\_\\'\\'"
		result := SqlString(s)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
	})
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "123ABCabc", SqlString("123ABCabc"))
	})
}
