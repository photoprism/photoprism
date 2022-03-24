package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlLike(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlLike(""))
	})
	t.Run("Special", func(t *testing.T) {
		s := "' \" \t \n %_''"
		exp := "\\' \\\"   %\\_\\'\\'"
		result := SqlLike(s)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
	})
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "123ABCabc", SqlLike("   123ABCabc%*    "))
	})
}
