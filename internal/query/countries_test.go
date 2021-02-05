package query

import (
	"testing"
)

func TestPurgeUnusedCountries(t *testing.T) {
	if err := PurgeUnusedCountries(); err != nil {
		t.Fatal(err)
	}
}
