package query

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSession(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result, err := Session("")
		t.Logf("session: %#v", result)
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.ID)
		assert.Equal(t, "", result.UserUID)
		assert.Equal(t, "", result.UserName)
	})
	t.Run("Invalid", func(t *testing.T) {
		result, err := Session("1234")
		assert.Error(t, err)
		assert.Equal(t, "invalid session id", err.Error())
		assert.ErrorIs(t, err, ErrInvalidSessionID)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.ID)
		assert.Equal(t, "", result.UserUID)
		assert.Equal(t, "", result.UserName)
	})
	t.Run("Alice", func(t *testing.T) {
		if result, err := Session("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("session: %#v", result)
			assert.NotNil(t, result)
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), result.ID)
			assert.Equal(t, "uqxetse3cy5eo9z2", result.UserUID)
			assert.Equal(t, "alice", result.UserName)
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if result, err := Session("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("session: %#v", result)
			assert.NotNil(t, result)
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"), result.ID)
			assert.Equal(t, "uqxc08w3d0ej2283", result.UserUID)
			assert.Equal(t, "bob", result.UserName)
		}
	})
}

func TestSessions(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		if results, err := Sessions(0, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			// t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Limit", func(t *testing.T) {
		if results, err := Sessions(1, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			// t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Offset", func(t *testing.T) {
		if results, err := Sessions(0, 1, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			// t.Logf("sessions: %#v", results)
		}
	})
	t.Run("SearchAlice", func(t *testing.T) {
		if results, err := Sessions(100, 0, "sess_expires DESC, user_name", "alice"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), results[0].ID)
				assert.Equal(t, "uqxetse3cy5eo9z2", results[0].UserUID)
				assert.Equal(t, "alice", results[0].UserName)
			}
		}
	})
	t.Run("SortByID", func(t *testing.T) {
		if results, err := Sessions(100, 0, "id", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			// t.Logf("sessions: %#v", results)
		}
	})
	t.Run("SearchAliceSortByID", func(t *testing.T) {
		if results, err := Sessions(100, 0, "id", "alice"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			// t.Logf("sessions: %#v", results)
		}
	})
	t.Run("SearchUpperAliceSortByID", func(t *testing.T) {
		if results, err := Sessions(100, 0, "id", "ALICE"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			// t.Logf("sessions: %#v", results)
		}
	})
}

func TestSession_ExactID(t *testing.T) {
	id := rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

	result, err := Session(id)
	require.NoError(t, err)
	assert.Equal(t, id, result.ID)

	_, err = Session(strings.ToUpper(id))
	assert.Error(t, err)
}

func TestSessions_Literal(t *testing.T) {
	base := "zzl" + rnd.Base36(5)

	for _, name := range []string{base + "_a", base + "Xa"} {
		s := entity.NewSession(3600, 0)
		s.UserName = name
		require.NoError(t, s.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(s).Error })
	}

	result, err := Sessions(100, 0, "", base+"_a")
	require.NoError(t, err)

	if assert.Len(t, result, 1) {
		assert.Equal(t, base+"_a", result[0].UserName)
	}
}
