package video

import "testing"

func TestFormats(t *testing.T) {
	if val := Formats[""]; val != AVC {
		t.Fatal("default type should be avc")
	}

	if val := Formats["mp4"]; val != MP4 {
		t.Fatal("mp4 type should be mp4")
	}

	if val := Formats["avc"]; val != AVC {
		t.Fatal("mp4 type should be avc")
	}
}
