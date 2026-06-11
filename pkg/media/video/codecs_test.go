package video

import "testing"

func TestCodecs(t *testing.T) {
	if val := Codecs[""]; val != CodecUnknown {
		t.Fatal("default codec should be CodecUnknown")
	}

	if val := Codecs["avc"]; val != CodecAvc1 {
		t.Fatal("codec should be CodecAVC")
	}

	if val := Codecs["av1"]; val != CodecAv01 {
		t.Fatal("codec should be CodecAV1")
	}

	if val := Codecs["evc"]; val != CodecEvc1 {
		t.Fatal("codec should be CodecEVC")
	}

	if val := Codecs["vvcC"]; val != CodecVvc1 {
		t.Fatal("codec should be CodecVVC")
	}

	if val := Codecs["magicyuv"]; val != CodecMagicYUV {
		t.Fatal("codec should be CodecMagicYUV")
	}

	if val := Codecs["h264"]; val != CodecAvc1 {
		t.Fatal("codec 'h264' should map to CodecAvc1")
	}

	if val := Codecs["h265"]; val != CodecHvc1 {
		t.Fatal("codec 'h265' should map to CodecHvc1")
	}

	for _, fourcc := range []string{"magy", "m8rg", "m8ra", "m8rb", "m8y0", "m8y2", "m8y4", "m8ya", "m8g0"} {
		if val := Codecs[fourcc]; val != CodecMagicYUV {
			t.Fatalf("FourCC %q should map to CodecMagicYUV, got %q", fourcc, val)
		}
	}

	for _, id := range []string{"vfw", "v_ms", "v_ms/vfw/fourcc"} {
		if val := Codecs[id]; val != CodecVFW {
			t.Fatalf("identifier %q should map to CodecVFW, got %q", id, val)
		}
	}
}

func TestCanonical(t *testing.T) {
	cases := map[string]Codec{
		"magy":            CodecMagicYUV,
		"magicyuv":        CodecMagicYUV, // Human-readable alias.
		"m8ra":            CodecMagicYUV, // FourCC reported by ExifTool.
		"h264":            CodecAvc1,
		"avc":             CodecAvc1,
		"h265":            CodecHvc1,
		"v_ms/vfw/fourcc": CodecVFW, // Matroska VFW wrapper.
		"v_ms":            CodecVFW,
		"mkv":             "mkv",    // Not a codec alias, returned unchanged.
		"a_opus":          "a_opus", // Maps to CodecUnknown, returned unchanged.
		"":                "",
	}

	for in, want := range cases {
		if got := Canonical(in); got != want {
			t.Fatalf("Canonical(%q) = %q, want %q", in, got, want)
		}
	}
}
