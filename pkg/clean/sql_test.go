package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlSpecial(t *testing.T) {
	t.Run("Special MySQL", func(t *testing.T) {
		if s, o := SqlSpecial(1, MySQL); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31, MySQL); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\\', MySQL); !s {
			t.Error("\\ is special")
		} else if o {
			t.Error("\\ must not be omitted")
		}

		if s, o := SqlSpecial('\'', MySQL); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})
	t.Run("Special SQLite", func(t *testing.T) {
		if s, o := SqlSpecial(1, SQLite3); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31, SQLite3); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\'', SQLite3); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})
	t.Run("Special Postgres", func(t *testing.T) {
		if s, o := SqlSpecial(1, Postgres); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31, Postgres); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\'', Postgres); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})

	t.Run("NotSpecial MySQL", func(t *testing.T) {
		if s, o := SqlSpecial(32, MySQL); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A', MySQL); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a', MySQL); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_', MySQL); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SqlSpecial('"', MySQL); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}
	})
	t.Run("NotSpecial SQLite", func(t *testing.T) {
		if s, o := SqlSpecial(32, SQLite3); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A', SQLite3); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a', SQLite3); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_', SQLite3); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SqlSpecial('"', SQLite3); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}

		if s, o := SqlSpecial('\\', SQLite3); s {
			t.Error("\\ is not special")
		} else if o {
			t.Error("\\ must not be omitted")
		}
	})
	t.Run("NotSpecial Postgres", func(t *testing.T) {
		if s, o := SqlSpecial(32, Postgres); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A', Postgres); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a', Postgres); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_', Postgres); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SqlSpecial('"', Postgres); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}

		if s, o := SqlSpecial('\\', Postgres); s {
			t.Error("\\ is not special")
		} else if o {
			t.Error("\\ must not be omitted")
		}
	})
}

func TestSqlString(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlString("", MySQL))
		assert.Equal(t, "", SqlString("", SQLite3))
		assert.Equal(t, "", SqlString("", Postgres))
	})
	t.Run("Special", func(t *testing.T) {
		s := "' \" \t \n %_''\\"
		exp := "'' \"   %_''''\\\\"
		result := SqlString(s, MySQL)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
		exp = "'' \"   %_''''\\"
		result = SqlString(s, SQLite3)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
		exp = "'' \"   %_''''\\"
		result = SqlString(s, Postgres)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
	})
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "123ABCabc", SqlString("123ABCabc", MySQL))
	})
}
