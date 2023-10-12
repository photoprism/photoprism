package video

import "testing"

func TestCodecs(t *testing.T) {
	if val := Codecs[""]; val != CodecUnknown {
		t.Fatal("default codec should be CodecUnknown")
	}

	if val := Codecs["avc"]; val != CodecAVC {
		t.Fatal("codec should be CodecAVC")
	}

	if val := Codecs["av1"]; val != CodecAV1 {
		t.Fatal("codec should be CodecAV1")
	}
}
