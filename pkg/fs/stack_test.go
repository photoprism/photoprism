package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStackPrefix verifies the stack names of capture files, their sidecars, and other files.
func TestStackPrefix(t *testing.T) {
	t.Run("Insta360Video", func(t *testing.T) {
		left := "VID_20220625_140410_00_008"

		for _, fileName := range []string{
			"VID_20220625_140410_00_008.insv",
			"VID_20220625_140410_10_008.insv",
			"LRV_20220625_140410_11_008.insv",
			"/originals/2022/VID_20220625_140410_10_008.insv",
			"VID_20220625_140410_10_008.INSV",
			"LRV_20220625_140410_11_008.Insv",
			"VID_20220625_140410_10_008.insv.jpg",
			"VID_20220625_140410_10_008.insv.avc",
			"LRV_20220625_140410_11_008.insv.jpg",
			"LRV_20220625_140410_11_008.insv.avc",
			"VID_20220625_140410_10_008.insv.xmp",
			"VID_20220625_140410_10_008.insv.jpg.json",
			"VID_20220625_140410_10_008.insv.unknownext",
		} {
			assert.Equal(t, left, StackPrefix(fileName, false), fileName)
			assert.Equal(t, left, StackPrefix(fileName, true), fileName)
		}
	})
	t.Run("Insta360Proxy", func(t *testing.T) {
		for _, fileName := range []string{
			"LRV_20240415_213145_01_035.lrv",
			"/originals/2024/LRV_20240415_213145_01_035.lrv",
			"LRV_20240415_213145_01_035.LRV",
			"LRV_20240415_213145_01_035.lrv.jpg",
			"LRV_20240415_213145_01_035.lrv.avc",
			"LRV_20240415_213145_01_035.lrv.xmp",
		} {
			assert.Equal(t, "VID_20240415_213145_00_035", StackPrefix(fileName, false), fileName)
			assert.Equal(t, "VID_20240415_213145_00_035", StackPrefix(fileName, true), fileName)
		}

		assert.Equal(t, "vid_20240415_213145_00_035", StackPrefix("lrv_20240415_213145_01_035.lrv", false))
	})
	t.Run("Insta360VideoLowercase", func(t *testing.T) {
		for _, fileName := range []string{
			"vid_20220625_140410_00_008.insv",
			"vid_20220625_140410_10_008.insv",
			"lrv_20220625_140410_11_008.insv",
			"lrv_20220625_140410_11_008.insv.jpg",
		} {
			assert.Equal(t, "vid_20220625_140410_00_008", StackPrefix(fileName, false), fileName)
			assert.Equal(t, "vid_20220625_140410_00_008", StackPrefix(fileName, true), fileName)
		}

		assert.Equal(t, "Vid_20220625_140410_00_008", StackPrefix("Vid_20220625_140410_10_008.insv", false))
		assert.Equal(t, "Vid_20220625_140410_00_008", StackPrefix("Lrv_20220625_140410_11_008.insv", false))
		assert.Equal(t, "vID_20220625_140410_00_008", StackPrefix("lRV_20220625_140410_11_008.insv", false))
	})

	// Names that no rule matches must return the same as BasePrefix.
	for _, fileName := range []string{
		"",
		"VID_20220625_140410_10_008.mp4",
		"VID_20220625_140410_10_008.jpg",
		"VID_20220625_140410_10_008.xmp",
		"VID_20220625_140410_10_008.MP4",
		"VID_20220625_140410_10_008_insv.mp4",
		"VID_20220625_140410_10_008.insvx",
		"VID_20220625_140410_10_008.mp4.insv",
		"VID_20220625_140410_10_008.insp",
		"IMG_20220625_140410_00_008.insp",
		"IMG_20220625_140410_10_008.insp",
		"IMG_20220625_140410_10_008.insp.jpg",
		"IMG_20220625_140410_10_008.insv",
		"IMG_20220625_140410_10_008.jpg",
		"IMG_20220625_140410_11_008.insp",
		"IMG_20220625_140410_01_008.insp",
		"IMG_2022_1404_10_8.insp",
		"IMG_20220625_140410_10_0008.insp",
		"copy-IMG_20220625_140410_10_008.insp",
		"IMG_20220625_140410_10_008 (2).insp",
		"IMG_20220625_140410_10_008.inspx",
		"LRV_20220625_140410_11_008.mp4",
		"LRV_20240415_213145_01_035.mp4",
		"LRV_20240415_213145_01_035.insv",
		"LRV_20240415_213145_00_035.lrv",
		"LRV_20240415_213145_11_035.lrv",
		"LRV_20240415_213145_10_035.lrv",
		"VID_20240415_213145_01_035.lrv",
		"VID_20240415_213145_00_035.lrv",
		"LRV_20240415_213145_01_35.lrv",
		"copy-LRV_20240415_213145_01_035.lrv",
		"LRV_20240415_213145_01_035 (2).lrv",
		"LRV_20240415_213145_01_035.lrvx",
		"proxy.lrv",
		"LRV_20220625_140410_10_008.insv",
		"LRV_20220625_140410_00_008.insv",
		"VID_20220625_140410_11_008.insv",
		"VID_20220625_140410_01_008.insv",
		"VID_20220625_140410_10_08.insv",
		"VID_20220625_140410_10_0008.insv",
		"VID_2022062_140410_10_008.insv",
		"copy-VID_20220625_140410_10_008.insv",
		"VID_20220625_140410_10_008 (2).insv",
		"VID_20220625_140410_10_008 copy.insv",
		"VID_20220625_140410_10_008.00001.insv",
		"VID_20220625_140410_10_008_01.insv",
		"VID_20220625_140410_10_008_01.mp4",
		"IMG_1234.jpg",
		"IMG_1234.mp4",
		"IMG_1234 (2).jpg",
		"/testdata/Test copy 3.jpg",
		"/testdata/Test.jpg.json",
		"Screenshot 2019-05-21 at 10.45.52.png",
		"20180506_091537_DSC02122.JPG",
		".insv",
		"insv",
		"VID_20220625_140410_10_008.in\u017fv",
		"VID_20220625_140410_10_008.in\u017fv.jpg",
	} {
		t.Run("Unchanged/"+fileName, func(t *testing.T) {
			assert.Equal(t, BasePrefix(fileName, false), StackPrefix(fileName, false))
			assert.Equal(t, BasePrefix(fileName, true), StackPrefix(fileName, true))
		})
	}
}

// TestInsta360StackName verifies role checks and prefix case handling for capture matches.
func TestInsta360StackName(t *testing.T) {
	t.Run("Left", func(t *testing.T) {
		assert.Equal(t, "VID_20220625_140410_00_008", insta360StackName([]string{"", "VID", "20220625", "140410", "00", "008"}))
	})
	t.Run("Proxy", func(t *testing.T) {
		assert.Equal(t, "VID_20220625_140410_00_008", insta360StackName([]string{"", "LRV", "20220625", "140410", "11", "008"}))
		assert.Equal(t, "vid_20220625_140410_00_008", insta360StackName([]string{"", "lrv", "20220625", "140410", "11", "008"}))
		assert.Equal(t, "VID_20240415_213145_00_035", insta360StackName([]string{"", "LRV", "20240415", "213145", "01", "035"}))
	})
	t.Run("InvalidRole", func(t *testing.T) {
		assert.Equal(t, "", insta360StackName([]string{"", "VID", "20220625", "140410", "11", "008"}))
		assert.Equal(t, "", insta360StackName([]string{"", "LRV", "20220625", "140410", "00", "008"}))
		assert.Equal(t, "", insta360StackName([]string{"", "IMG", "20220625", "140410", "10", "008"}))
		assert.Equal(t, "", insta360StackName([]string{"", "DSC", "20220625", "140410", "00", "008"}))
		assert.Equal(t, "", insta360StackName([]string{"", "VID", "20220625", "140410", "01", "008"}))
		assert.Equal(t, "", insta360StackName([]string{"", "IMG", "20220625", "140410", "01", "008"}))
	})
	t.Run("InvalidMatch", func(t *testing.T) {
		assert.Equal(t, "", insta360StackName(nil))
		assert.Equal(t, "", insta360StackName([]string{"VID_20220625_140410_00_008.insv"}))
	})
}

// TestStackRuleName verifies that rules only match at an extension boundary.
func TestStackRuleName(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		assert.Equal(t, "VID_20220625_140410_00_008", stackRuleName("VID_20220625_140410_10_008.insv", "VID_20220625_140410_10_008"))
		assert.Equal(t, "VID_20220625_140410_00_008", stackRuleName("VID_20220625_140410_10_008.insv.jpg", "VID_20220625_140410_10_008"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.Equal(t, "", stackRuleName("VID_20220625_140410_10_008.insvx", "VID_20220625_140410_10_008"))
		assert.Equal(t, "", stackRuleName("VID_20220625_140410_10_008", "VID_20220625_140410_10_008"))
		assert.Equal(t, "", stackRuleName("", ""))
	})
}

// TestMatchCase verifies that letters follow the case of the reference at the same position.
func TestMatchCase(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "VID", matchCase("VID", "LRV"))
		assert.Equal(t, "vid", matchCase("VID", "lrv"))
		assert.Equal(t, "Vid", matchCase("VID", "Lrv"))
		assert.Equal(t, "vId", matchCase("VID", "lRv"))
	})
	t.Run("ShortReference", func(t *testing.T) {
		assert.Equal(t, "vID", matchCase("VID", "l"))
		assert.Equal(t, "VID", matchCase("VID", ""))
	})
}

// TestInsta360Patterns verifies that the capture patterns match complete file names only.
func TestInsta360Patterns(t *testing.T) {
	t.Run("Video", func(t *testing.T) {
		assert.True(t, Insta360VideoPattern.MatchString("VID_20220625_140410_10_008.insv"))
		assert.True(t, Insta360VideoPattern.MatchString("lrv_20220625_140410_11_008.INSV"))
		assert.False(t, Insta360VideoPattern.MatchString("VID_20220625_140410_10_008.insv.jpg"))
		assert.False(t, Insta360VideoPattern.MatchString("copy-VID_20220625_140410_10_008.insv"))
		assert.False(t, Insta360VideoPattern.MatchString("VID_20220625_140410_10_008.insp"))
		assert.False(t, Insta360VideoPattern.MatchString("VID_20220625_140410_10_008.in\u017fv"))
		assert.False(t, Insta360VideoPattern.MatchString("LRV_20240415_213145_01_035.insv"))
	})
	t.Run("Proxy", func(t *testing.T) {
		assert.True(t, Insta360ProxyPattern.MatchString("LRV_20240415_213145_01_035.lrv"))
		assert.True(t, Insta360ProxyPattern.MatchString("lrv_20240415_213145_01_035.LRV"))
		assert.False(t, Insta360ProxyPattern.MatchString("LRV_20240415_213145_01_035.lrv.jpg"))
		assert.False(t, Insta360ProxyPattern.MatchString("LRV_20240415_213145_11_035.lrv"))
		assert.False(t, Insta360ProxyPattern.MatchString("VID_20240415_213145_01_035.lrv"))
		assert.False(t, Insta360ProxyPattern.MatchString("LRV_20240415_213145_01_035.insv"))
	})
}

// TestKeepStacked verifies that only the lens and proxy originals of a capture must stay stacked.
func TestKeepStacked(t *testing.T) {
	t.Run("Capture", func(t *testing.T) {
		for _, fileName := range []string{
			"VID_20220625_140410_10_008.insv",
			"LRV_20220625_140410_11_008.insv",
			"/originals/2022/VID_20220625_140410_10_008.INSV",
			"vid_20220625_140410_10_008.insv",
			"LRV_20240415_213145_01_035.lrv",
		} {
			assert.True(t, KeepStacked(fileName), fileName)
		}
	})
	t.Run("Other", func(t *testing.T) {
		// Left lens files are named like single-file captures, so they are not flagged by name.
		for _, fileName := range []string{
			"",
			"VID_20220625_140410_00_008.insv",
			"IMG_20220625_140410_00_008.insp",
			"IMG_20220625_140410_10_008.insp",
			"IMG_20231015_101112_00_123.insp",
			"VID_20220625_140410_10_008.insv.jpg",
			"VID_20220625_140410_10_008.mp4",
			"VID_20220625_140410_11_008.insv",
			"LRV_20220625_140410_00_008.insv",
			"IMG_20220625_140410_11_008.insp",
			"IMG_20220625_140410_10_008.jpg",
			"VID_20220625_140410_10_008 (2).insv",
			"insta360.insv",
			"LRV_20240415_213145_01_035.lrv.jpg",
			"LRV_20240415_213145_11_035.lrv",
			"IMG_1234.jpg",
		} {
			assert.False(t, KeepStacked(fileName), fileName)
		}
	})
}

// TestStackGroup verifies the shared stack name of capture originals.
func TestStackGroup(t *testing.T) {
	t.Run("Capture", func(t *testing.T) {
		for _, fileName := range []string{
			"VID_20220625_140410_00_008.insv",
			"VID_20220625_140410_10_008.insv",
			"/originals/LRV_20220625_140410_11_008.insv",
		} {
			assert.Equal(t, "VID_20220625_140410_00_008", StackGroup(fileName), fileName)
		}

		assert.Equal(t, "VID_20240415_213145_00_035", StackGroup("LRV_20240415_213145_01_035.lrv"))

	})
	t.Run("Other", func(t *testing.T) {
		for _, fileName := range []string{
			"",
			"VID_20220625_140410_10_008.insv.jpg",
			"VID_20220625_140410_10_008.mp4",
			"VID_20220625_140410_11_008.insv",
			"IMG_20231015_101112_00_123.insp",
			"IMG_20231015_101112_10_124.insp",
			"IMG_1234.jpg",
		} {
			assert.Equal(t, "", StackGroup(fileName), fileName)
		}
	})
}
