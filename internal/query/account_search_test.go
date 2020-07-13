package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccounts(t *testing.T) {
	t.Run("find accounts", func(t *testing.T) {
		f := form.AccountSearch{
			Query:  "",
			Share:  true,
			Sync:   true,
			Status: "",
			Count:  10,
			Offset: 0,
			Order:  "",
		}
		r, err := AccountSearch(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("accounts: %+v", r)

		assert.LessOrEqual(t, 1, len(r))

		for _, r := range r {
			assert.IsType(t, entity.Account{}, r)
		}
	})

	t.Run("find accounts count 1001", func(t *testing.T) {
		f := form.AccountSearch{
			Query:  "",
			Share:  false,
			Sync:   false,
			Status: "test",
			Count:  1001,
			Offset: 0,
			Order:  "",
		}
		r, err := AccountSearch(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("accounts: %+v", r)

		assert.LessOrEqual(t, 1, len(r))

		for _, r := range r {
			assert.IsType(t, entity.Account{}, r)
		}
	})
	t.Run("find accounts count > max results", func(t *testing.T) {
		f := form.AccountSearch{
			Query:  "",
			Status: "test",
			Count:  100000,
			Offset: 0,
			Order:  "",
		}
		r, err := AccountSearch(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("accounts: %+v", r)

		assert.LessOrEqual(t, 1, len(r))

		for _, r := range r {
			assert.IsType(t, entity.Account{}, r)
		}
	})
}

func TestAccountByID(t *testing.T) {
	t.Run("existing account", func(t *testing.T) {
		r, err := AccountByID(uint(1000001))

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Test Account2", r.AccName)

	})
	t.Run("record not found", func(t *testing.T) {
		r, err := AccountByID(uint(123))

		if err == nil {
			t.Fatal()
		}
		assert.Equal(t, "record not found", err.Error())
		assert.Empty(t, r)
	})
}
