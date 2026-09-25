package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClients(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		if results, err := Clients(0, 0, "", "", false); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 4, len(results))
		}
	})
	t.Run("Limit", func(t *testing.T) {
		if results, err := Clients(2, 0, "", "", false); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
		}
	})
	t.Run("Offset", func(t *testing.T) {
		if results, err := Clients(3, 1, "", "all", false); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 3, len(results))
		}
	})
	t.Run("SearchAliceByName", func(t *testing.T) {
		if results, err := Clients(100, 0, "", "alice", false); err != nil {
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
	t.Run("SearchAliceByNameUppercase", func(t *testing.T) {
		if results, err := Clients(100, 0, "", "ALICE", false); err != nil {
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
		if results, err := Clients(100, 0, "", "cs5gfen1bgxz7s9i", false); err != nil {
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
		if results, err := Clients(100, 0, "", "uqxetse3cy5eo9z2", false); err != nil {
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
		if results, err := Clients(100, 0, "created_at", "", false); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 4, len(results))
		}
	})
}

func TestClients_Literal(t *testing.T) {
	base := "zzl" + rnd.Base36(5)

	for _, name := range []string{base + "_a", base + "Xa"} {
		c := entity.NewClient()
		c.SetName(name)
		require.NoError(t, c.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(c).Error })
	}

	result, err := Clients(100, 0, "", base+"_a", false)
	require.NoError(t, err)

	if assert.Len(t, result, 1) {
		assert.Equal(t, base+"_a", result[0].ClientName)
	}
}
