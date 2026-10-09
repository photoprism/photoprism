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
	t.Run("GooglePixelPhoto", func(t *testing.T) {
		canonical := "PXL_20230805_123456789"

		for _, fileName := range []string{
			// Standard Photo RAW+JPEG
			"PXL_20230805_123456789.RAW-01.jpg",
			"PXL_20230805_123456789.RAW-01.COVER.jpg",
			"PXL_20230805_123456789.RAW-02.ORIGINAL.dng",
			// Motion Photo RAW+JPEG
			"PXL_20230805_123456789.RAW-01.MP.jpg",
			"PXL_20230805_123456789.RAW-01.MP.COVER.jpg",
			// Night Sight Photo RAW+JPEG
			"PXL_20230805_123456789.NIGHT.RAW-01.jpg",
			"PXL_20230805_123456789.NIGHT.RAW-01.COVER.jpg",
			"PXL_20230805_123456789.NIGHT.RAW-02.ORIGINAL.dng",
			// Portrait Photo JPEG
			"PXL_20230805_123456789.PORTRAIT.jpg",
			"PXL_20230805_123456789.PORTRAIT.ORIGINAL.jpg",
			"PXL_20230805_123456789.PORTRAIT-01.COVER.jpg",
			"PXL_20230805_123456789.PORTRAIT-02.ORIGINAL.jpg",
			// Add Me Photo JPEG
			"PXL_20230805_123456789.BURST-01.jpg",
			"PXL_20230805_123456789.BURST-02.jpg",
			"PXL_20230805_123456789.BURST-03.jpg",
			// Long Exposure Photo JPEG
			"PXL_20230805_123456789.LONG_EXPOSURE-01.jpg",
			"PXL_20230805_123456789.LONG_EXPOSURE-01.COVER.jpg",
			"PXL_20230805_123456789.LONG_EXPOSURE-02.ORIGINAL.jpg",
			// Action Pan Photo JPEG
			"PXL_20230805_123456789.ACTION_PAN-01.jpg",
			"PXL_20230805_123456789.ACTION_PAN-01.COVER.jpg",
			"PXL_20230805_123456789.ACTION_PAN-02.ORIGINAL.jpg",
			// AI Pro Zoom Photo RAW+JPEG
			"PXL_20230805_123456789.BURST-02.original.jpg",
			"PXL_20230805_123456789.BURST-03.ORIGINAL.dng",
			// Paths and Sidecars
			"/originals/2023/08/PXL_20230805_123456789.RAW-01.COVER.jpg",
			"/originals/2023/08/PXL_20230805_123456789.RAW-02.ORIGINAL.dng",
			"PXL_20230805_123456789.RAW-01.COVER.jpg.xmp",
			"PXL_20230805_123456789.RAW-01.COVER.xmp",
			"PXL_20230805_123456789.RAW-01.COVER.jpg.json",
		} {
			assert.Equal(t, canonical, StackPrefix(fileName, false), fileName)
			assert.Equal(t, canonical, StackPrefix(fileName, true), fileName)
		}
	})
	t.Run("GooglePixelVideo", func(t *testing.T) {
		canonical := "PXL_20230805_123456789"

		for _, fileName := range []string{
			// Video Boost MP4
			"PXL_20230805_123456789.VB-01.COVER.mp4",
			"PXL_20230805_123456789.VB-02.MAIN.mp4",
			"PXL_20230805_123456789.VB-03.MAIN.mp4",
			// Night Sight Video MP4
			"PXL_20230805_123456789.NS-01.COVER.mp4",
			"PXL_20230805_123456789.NS-02.MAIN.mp4",
			"PXL_20230805_123456789.NS-03.MAIN.mp4",
			// Paths and Sidecars
			"/originals/2023/08/PXL_20230805_123456789.VB-02.MAIN.mp4",
			"PXL_20230805_123456789.VB-02.MAIN.mp4.xmp",
			"PXL_20230805_123456789.VB-02.MAIN.xmp",
			"PXL_20230805_123456789.VB-02.MAIN.mp4.json",
		} {
			assert.Equal(t, canonical, StackPrefix(fileName, false), fileName)
			assert.Equal(t, canonical, StackPrefix(fileName, true), fileName)
		}
	})
	t.Run("GooglePixelLowercase", func(t *testing.T) {
		for _, fileName := range []string{
			"pxl_20230805_123456789.raw-01.cover.jpg",
			"pxl_20230805_123456789.raw-02.original.dng",
		} {
			assert.Equal(t, "pxl_20230805_123456789", StackPrefix(fileName, false), fileName)
			assert.Equal(t, "pxl_20230805_123456789", StackPrefix(fileName, true), fileName)
		}

		assert.Equal(t, "Pxl_20230805_123456789", StackPrefix("Pxl_20230805_123456789.raw-01.cover.jpg", false))
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
		"PXL_20230805_123456.jpg",
		"PXL_20230805_123456.MP.jpg",
		"PXL_20230805_123456.NIGHT.jpg",
		"PXL_20230805_123456.PANO.jpg",
		"PXL_20230805_123456.PHOTOSPHERE.jpg",
		"PXL_20230805_123456.TS.mp4",
		"PXL_20230805_123456.SLOW_MOTION.mp4",
		"PXL_20230805_123456.CINEMATIC.mp4",
		"IMG_20230805_123456789.RAW-01.COVER.jpg",
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

// TestGooglePixelStackName verifies canonical base prefix extraction for Google Pixel Camera captures.
func TestGooglePixelStackName(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		assert.Equal(t, "PXL_20230805_123456789", googlePixelStackName("PXL_20230805_123456789.RAW-01.COVER"))
		assert.Equal(t, "PXL_20230805_123456789", googlePixelStackName("PXL_20230805_123456789.RAW-02.ORIGINAL"))
		assert.Equal(t, "PXL_20230805_123456789", googlePixelStackName("PXL_20230805_123456789.PORTRAIT-01.COVER"))
		assert.Equal(t, "PXL_20230805_123456789", googlePixelStackName("PXL_20230805_123456789.VB-02.MAIN"))
		assert.Equal(t, "pxl_20230805_123456789", googlePixelStackName("pxl_20230805_123456789.raw-01.cover"))
		assert.Equal(t, "Pxl_20230805_123456789", googlePixelStackName("Pxl_20230805_123456789.raw-01.cover"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.Equal(t, "", googlePixelStackName("PXL_20230805_123456"))
		assert.Equal(t, "", googlePixelStackName("PXL_20230805_123456.MP"))
		assert.Equal(t, "", googlePixelStackName("IMG_20230805_123456789.RAW-01.COVER"))
		assert.Equal(t, "", googlePixelStackName(""))
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

// TestGooglePixelPattern verifies that the capture pattern matches Google Pixel Camera multi-file capture prefixes.
func TestGooglePixelPattern(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		assert.True(t, GooglePixelPattern.MatchString("PXL_20230805_123456789.RAW-01.COVER"))
		assert.True(t, GooglePixelPattern.MatchString("PXL_20230805_123456789.RAW-02.ORIGINAL"))
		assert.True(t, GooglePixelPattern.MatchString("PXL_20230805_123456789.PORTRAIT.ORIGINAL"))
		assert.True(t, GooglePixelPattern.MatchString("PXL_20230805_123456789.PORTRAIT"))
		assert.True(t, GooglePixelPattern.MatchString("PXL_20230805_123456789.VB-01.COVER"))
		assert.True(t, GooglePixelPattern.MatchString("PXL_20230805_123456789.VB-02.MAIN"))
		assert.True(t, GooglePixelPattern.MatchString("pxl_20230805_123456789.raw-01.cover"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.False(t, GooglePixelPattern.MatchString("PXL_20230805_123456"))
		assert.False(t, GooglePixelPattern.MatchString("PXL_20230805_123456.MP"))
		assert.False(t, GooglePixelPattern.MatchString("IMG_20230805_123456789.RAW-01.COVER"))
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
