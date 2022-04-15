package video

import "testing"

func TestTypes(t *testing.T) {
	if val := Types[""]; val != AVC {
		t.Fatal("default type should be avc")
	}

	if val := Types["mp4"]; val != MP4 {
		t.Fatal("mp4 type should be mp4")
	}

	if val := Types["avc"]; val != AVC {
		t.Fatal("mp4 type should be avc")
	}
}
