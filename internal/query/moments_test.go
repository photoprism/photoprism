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
			assert.Len(t, moment.PhotoCountry, 0)
			assert.GreaterOrEqual(t, moment.PhotoYear, 1990)
			assert.LessOrEqual(t, moment.PhotoYear, 2800)
			assert.GreaterOrEqual(t, moment.PhotoMonth, 1)
			assert.LessOrEqual(t, moment.PhotoMonth, 12)
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
			assert.Len(t, moment.PhotoCountry, 2)
			assert.GreaterOrEqual(t, moment.PhotoYear, 1990)
			assert.LessOrEqual(t, moment.PhotoYear, 2800)
			assert.Equal(t, moment.PhotoMonth, 0)
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
			assert.Len(t, moment.PhotoCountry, 2)
			assert.NotEmpty(t, moment.PhotoState)
			assert.Equal(t, moment.PhotoYear, 0)
			assert.Equal(t, moment.PhotoMonth, 0)
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
			assert.NotEmpty(t, moment.PhotoCategory)
			assert.Empty(t, moment.PhotoCountry)
			assert.Empty(t, moment.PhotoState)
			assert.Equal(t, moment.PhotoYear, 0)
			assert.Equal(t, moment.PhotoMonth, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
		}
	})
}
