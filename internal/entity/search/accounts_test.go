package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestAccounts(t *testing.T) {
	t.Run("FindAccounts", func(t *testing.T) {
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
	t.Run("FindAccountsCountNum1001", func(t *testing.T) {
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
	t.Run("FindAccountsCountGreaterThanMaxResults", func(t *testing.T) {
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
	t.Run("QueryMatchesAccountName", func(t *testing.T) {
		f := form.SearchServices{
			Query: "Account2",
			Count: 10,
		}
		r, err := Accounts(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, r)

		found := false
		for _, s := range r {
			assert.Contains(t, s.AccName, "Account2")
			if s.AccName == "Test Account2" {
				found = true
			}
		}
		assert.True(t, found)
	})
	t.Run("QueryMatchesMultiple", func(t *testing.T) {
		f := form.SearchServices{
			Query: "Test Account",
			Count: 10,
		}
		r, err := Accounts(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(r), 2)
	})
	t.Run("QueryNoMatch", func(t *testing.T) {
		f := form.SearchServices{
			Query: "NoSuchServiceName",
			Count: 10,
		}
		r, err := Accounts(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, r)
	})
}
