package video

import "testing"

func TestTypes(t *testing.T) {
	if val := Types[""]; val != TypeMP4 {
		t.Fatal("default type should be TypeMP4")
	}

	if val := Types["mp4"]; val != TypeMP4 {
		t.Fatal("mp4 type should be TypeMP4")
	}
}
