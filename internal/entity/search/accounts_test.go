package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestAccounts_Bindable(t *testing.T) {
	expected, err := Accounts(form.SearchServices{Query: "Test", Count: 10})
	require.NoError(t, err)
	require.NotEmpty(t, expected)

	result, err := Accounts(form.SearchServices{Query: "Te\x00st\xff", Count: 10})
	require.NoError(t, err)
	assert.Len(t, result, len(expected))

	for _, status := range []string{entity.SyncStatusRefresh, entity.SyncStatusSynced} {
		expected, err = Accounts(form.SearchServices{Status: status, Count: 10})
		require.NoError(t, err)

		result, err = Accounts(form.SearchServices{Status: status + "\x00", Count: 10})
		require.NoError(t, err)
		assert.Len(t, result, len(expected), status)
	}
}

func TestAccounts_NoMatch(t *testing.T) {
	// A filter value of only NUL bytes and invalid UTF-8 matches nothing, while an empty one stays unfiltered.
	all, err := Accounts(form.SearchServices{Count: 10})
	require.NoError(t, err)
	require.NotEmpty(t, all)

	for _, frm := range []form.SearchServices{{Query: "\x00", Count: 10}, {Query: "\xff\xfe", Count: 10}, {Status: "\xff", Count: 10}, {Query: "\x00 ", Count: 10}, {Query: "\xff*", Count: 10}} {
		result, err := Accounts(frm)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result, "%q", frm.Query+frm.Status)
	}
}
