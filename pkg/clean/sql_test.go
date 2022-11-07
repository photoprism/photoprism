package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlSpecial(t *testing.T) {
	t.Run("Special", func(t *testing.T) {
		if s, o := SqlSpecial(1); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\\'); !s {
			t.Error("\\ is special")
		} else if o {
			t.Error("\\ must not be omitted")
		}

		if s, o := SqlSpecial('\''); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}

		if s, o := SqlSpecial('"'); !s {
			t.Error("\" is special")
		} else if o {
			t.Error("\" must not be omitted")
		}
	})
	t.Run("NotSpecial", func(t *testing.T) {
		if s, o := SqlSpecial(32); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A'); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a'); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_'); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}
	})
}

func TestSqlString(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlString(""))
	})
	t.Run("Special", func(t *testing.T) {
		s := "' \" \t \n %_''"
		exp := "'' \"\"   %_''''"
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
