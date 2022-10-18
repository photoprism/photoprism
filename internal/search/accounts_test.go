package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestAccounts(t *testing.T) {
	t.Run("find accounts", func(t *testing.T) {
		f := form.SearchServices{
			Query:  "",
			Share:  true,
			Sync:   true,
			Status: "",
			Count:  10,
			Offset: 0,
			Order:  "",
		}
		r, err := Accounts(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("accounts: %+v", r)

		assert.LessOrEqual(t, 1, len(r))

		for _, r := range r {
			assert.IsType(t, entity.Service{}, r)
		}
	})

	t.Run("find accounts count 1001", func(t *testing.T) {
		f := form.SearchServices{
			Query:  "",
			Share:  false,
			Sync:   false,
			Status: "refresh",
			Count:  1001,
			Offset: 0,
			Order:  "",
		}
		r, err := Accounts(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("accounts: %+v", r)

		assert.LessOrEqual(t, 1, len(r))

		for _, r := range r {
			assert.IsType(t, entity.Service{}, r)
		}
	})
	t.Run("find accounts count > max results", func(t *testing.T) {
		f := form.SearchServices{
			Query:  "",
			Status: "refresh",
			Count:  100000,
			Offset: 0,
			Order:  "",
		}
		r, err := Accounts(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("accounts: %+v", r)

		assert.LessOrEqual(t, 1, len(r))

		for _, r := range r {
			assert.IsType(t, entity.Service{}, r)
		}
	})
}
