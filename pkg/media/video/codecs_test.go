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

	if val := Codecs["evc"]; val != CodecEVC {
		t.Fatal("codec should be CodecEVC")
	}

	if val := Codecs["vvcC"]; val != CodecVVC {
		t.Fatal("codec should be CodecVVC")
	}
}
