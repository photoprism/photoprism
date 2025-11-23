package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/enum"
)

func TestSQLSpecial(t *testing.T) {
	t.Run("Special MySQL", func(t *testing.T) {
		if s, o := SQLSpecial(1, enum.MySQL); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SQLSpecial(31, enum.MySQL); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SQLSpecial('\\', enum.MySQL); !s {
			t.Error("\\ is special")
		} else if o {
			t.Error("\\ must not be omitted")
		}

		if s, o := SQLSpecial('\'', enum.MySQL); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})
	t.Run("Special SQLite", func(t *testing.T) {
		if s, o := SQLSpecial(1, enum.SQLite3); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SQLSpecial(31, enum.SQLite3); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SQLSpecial('\'', enum.SQLite3); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})
	t.Run("Special Postgres", func(t *testing.T) {
		if s, o := SQLSpecial(1, enum.Postgres); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SQLSpecial(31, enum.Postgres); !s {
			t.Error("char is special")
		} else if !o {
			t.Error("\" must be omitted")
		}

		if s, o := SQLSpecial('\'', enum.Postgres); !s {
			t.Error("' is special")
		} else if o {
			t.Error("' must not be omitted")
		}
	})

	t.Run("NotSpecial MySQL", func(t *testing.T) {
		if s, o := SQLSpecial(32, enum.MySQL); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SQLSpecial('A', enum.MySQL); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SQLSpecial('a', enum.MySQL); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SQLSpecial('_', enum.MySQL); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SQLSpecial('"', enum.MySQL); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}
	})
	t.Run("NotSpecial SQLite", func(t *testing.T) {
		if s, o := SQLSpecial(32, enum.SQLite3); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SQLSpecial('A', enum.SQLite3); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SQLSpecial('a', enum.SQLite3); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SQLSpecial('_', enum.SQLite3); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SQLSpecial('"', enum.SQLite3); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}

		if s, o := SQLSpecial('\\', enum.SQLite3); s {
			t.Error("\\ is not special")
		} else if o {
			t.Error("\\ must not be omitted")
		}
	})
	t.Run("NotSpecial Postgres", func(t *testing.T) {
		if s, o := SQLSpecial(32, enum.Postgres); s {
			t.Error("space is not special")
		} else if o {
			t.Error("space must not be omitted")
		}

		if s, o := SQLSpecial('A', enum.Postgres); s {
			t.Error("A is not special")
		} else if o {
			t.Error("A must not be omitted")
		}

		if s, o := SQLSpecial('a', enum.Postgres); s {
			t.Error("a is not special")
		} else if o {
			t.Error("a must not be omitted")
		}

		if s, o := SQLSpecial('_', enum.Postgres); s {
			t.Error("_ is not special")
		} else if o {
			t.Error("_ must not be omitted")
		}

		if s, o := SQLSpecial('"', enum.Postgres); s {
			t.Error("\" is not special")
		} else if o {
			t.Error("\" must not be omitted")
		}

		if s, o := SQLSpecial('\\', enum.Postgres); s {
			t.Error("\\ is not special")
		} else if o {
			t.Error("\\ must not be omitted")
		}
	})
}

func TestSqlString(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SQLString("", enum.MySQL))
		assert.Equal(t, "", SQLString("", enum.SQLite3))
		assert.Equal(t, "", SQLString("", enum.Postgres))
	})
	t.Run("Special", func(t *testing.T) {
		s := "' \" \t \n %_''\\"
		exp := "'' \"   %_''''\\\\"
		result := SQLString(s, enum.MySQL)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
		exp = "'' \"   %_''''\\"
		result = SQLString(s, enum.SQLite3)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
		exp = "'' \"   %_''''\\"
		result = SQLString(s, enum.Postgres)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
	})
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "123ABCabc", SQLString("123ABCabc", enum.MySQL))
	})
}
