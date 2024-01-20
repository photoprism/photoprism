package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClients(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		if results, err := Clients(0, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 4, len(results))
		}
	})
	t.Run("Limit", func(t *testing.T) {
		if results, err := Clients(2, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
		}
	})
	t.Run("Offset", func(t *testing.T) {
		if results, err := Clients(3, 1, "", "all"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 3, len(results))
		}
	})
	t.Run("SearchAliceByName", func(t *testing.T) {
		if results, err := Clients(100, 0, "", "alice"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "cs5gfen1bgxz7s9i", results[0].ClientUID)
				assert.Equal(t, "uqxetse3cy5eo9z2", results[0].UserUID)
				assert.Equal(t, "alice", results[0].UserName)
			}
		}
	})
	t.Run("SearchAliceByClientUID", func(t *testing.T) {
		if results, err := Clients(100, 0, "", "cs5gfen1bgxz7s9i"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "cs5gfen1bgxz7s9i", results[0].ClientUID)
				assert.Equal(t, "uqxetse3cy5eo9z2", results[0].UserUID)
				assert.Equal(t, "alice", results[0].UserName)
			}
		}
	})
	t.Run("SearchAliceByUserUID", func(t *testing.T) {
		if results, err := Clients(100, 0, "", "uqxetse3cy5eo9z2"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "cs5gfen1bgxz7s9i", results[0].ClientUID)
				assert.Equal(t, "uqxetse3cy5eo9z2", results[0].UserUID)
				assert.Equal(t, "alice", results[0].UserName)
			}
		}
	})
	t.Run("SortByCreated", func(t *testing.T) {
		if results, err := Clients(100, 0, "created_at", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 4, len(results))
		}
	})
}
