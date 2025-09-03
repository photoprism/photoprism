package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/constants"
)

func TestSqlSpecial(t *testing.T) {
	t.Run("Special MySQL", func(t *testing.T) {
		if s, o := SqlSpecial(1, constants.MySQL); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31, constants.MySQL); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\\', constants.MySQL); !s {
			t.Error("\\ is special")
		} else if o {
			t.Error("\\ must not be omitted")
		}

		if s, o := SqlSpecial('\'', constants.MySQL); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})
	t.Run("Special SQLite", func(t *testing.T) {
		if s, o := SqlSpecial(1, constants.SQLite3); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31, constants.SQLite3); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\'', constants.SQLite3); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})
	t.Run("Special Postgres", func(t *testing.T) {
		if s, o := SqlSpecial(1, constants.Postgres); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial(31, constants.Postgres); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SqlSpecial('\'', constants.Postgres); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})

	t.Run("NotSpecial MySQL", func(t *testing.T) {
		if s, o := SqlSpecial(32, constants.MySQL); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A', constants.MySQL); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a', constants.MySQL); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_', constants.MySQL); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SqlSpecial('"', constants.MySQL); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}
	})
	t.Run("NotSpecial SQLite", func(t *testing.T) {
		if s, o := SqlSpecial(32, constants.SQLite3); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A', constants.SQLite3); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a', constants.SQLite3); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_', constants.SQLite3); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SqlSpecial('"', constants.SQLite3); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}

		if s, o := SqlSpecial('\\', constants.SQLite3); s {
			t.Error("\\ is not special")
		} else if o {
			t.Error("\\ must not be omitted")
		}
	})
	t.Run("NotSpecial Postgres", func(t *testing.T) {
		if s, o := SqlSpecial(32, constants.Postgres); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SqlSpecial('A', constants.Postgres); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SqlSpecial('a', constants.Postgres); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SqlSpecial('_', constants.Postgres); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SqlSpecial('"', constants.Postgres); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}

		if s, o := SqlSpecial('\\', constants.Postgres); s {
			t.Error("\\ is not special")
		} else if o {
			t.Error("\\ must not be omitted")
		}
	})
}

func TestSqlString(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlString("", constants.MySQL))
		assert.Equal(t, "", SqlString("", constants.SQLite3))
		assert.Equal(t, "", SqlString("", constants.Postgres))
	})
	t.Run("Special", func(t *testing.T) {
		s := "' \" \t \n %_''\\"
		exp := "'' \"   %_''''\\\\"
		result := SqlString(s, constants.MySQL)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
		exp = "'' \"   %_''''\\"
		result = SqlString(s, constants.SQLite3)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
		exp = "'' \"   %_''''\\"
		result = SqlString(s, constants.Postgres)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
	})
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "123ABCabc", SqlString("123ABCabc", constants.MySQL))
	})
}
