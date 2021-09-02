package entity

import "testing"

func TestFile_MarshalJSON(t *testing.T) {
	if m := FileFixtures.Pointer("Video.mp4"); m == nil {
		t.Fatal("must not be nil")
	} else if j, err := m.MarshalJSON(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("json: %s", j)
	}
}
