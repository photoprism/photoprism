package entity

import (
	"testing"
)

func TestMarker_MarshalJSON(t *testing.T) {
	if m := MarkerFixtures.Pointer("actor-a-2"); m == nil {
		t.Fatal("must not be nil")
	} else if j, err := m.MarshalJSON(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("json: %s", j)
	}
}
