package entity

import (
	"testing"
)

func TestUpdatePhotoCounts(t *testing.T) {
	err := UpdatePhotoCounts()

	if err != nil {
		t.Fatal(err)
	}
}
