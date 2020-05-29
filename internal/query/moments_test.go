package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMomentsTime(t *testing.T) {
	t.Run("result found", func(t *testing.T) {
		results, err := MomentsTime(1)

		if err != nil {
			t.Fatal(err)
		}
		if len(results) < 4 {
			t.Error("at least 4 results expected")
		}

		t.Logf("MomentsTime %+v", results)

		for _, moment := range results {
			assert.Len(t, moment.Country, 0)
			assert.GreaterOrEqual(t, moment.Year, 1990)
			assert.LessOrEqual(t, moment.Year, 2800)
			assert.GreaterOrEqual(t, moment.Month, 1)
			assert.LessOrEqual(t, moment.Month, 12)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
		}
	})
}

func TestMomentsCountries(t *testing.T) {
	t.Run("result found", func(t *testing.T) {
		results, err := MomentsCountries(1)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsCountries %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			assert.Len(t, moment.Country, 2)
			assert.GreaterOrEqual(t, moment.Year, 1990)
			assert.LessOrEqual(t, moment.Year, 2800)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
		}
	})
}

func TestMomentsStates(t *testing.T) {
	t.Run("result found", func(t *testing.T) {
		results, err := MomentsStates(1)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsStates %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			assert.Len(t, moment.Country, 2)
			assert.NotEmpty(t, moment.State)
			assert.Equal(t, moment.Year, 0)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
		}
	})
}

func TestMomentsCategories(t *testing.T) {
	t.Run("result found", func(t *testing.T) {
		results, err := MomentsCategories(1)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsCategories %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			assert.NotEmpty(t, moment.Category)
			assert.Empty(t, moment.Country)
			assert.Empty(t, moment.State)
			assert.Equal(t, moment.Year, 0)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
		}
	})
}
