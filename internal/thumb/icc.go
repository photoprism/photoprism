package thumb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Standard ICC profiles located in "assets/profiles/icc".
const (
	// IccAdobeRGBCompat is compatible with Adobe RGB (1998).
	IccAdobeRGBCompat = "a98.icc"

	// IccAdobeRGBCompatV2 is A98C (Adobe RGB 1998 compatible, ICC v2).
	IccAdobeRGBCompatV2 = "adobecompat-v2.icc"
	// IccAdobeRGBCompatV4 is A98C (Adobe RGB 1998 compatible, ICC v4).
	IccAdobeRGBCompatV4 = "adobecompat-v4.icc"

	// IccAppleCompatV2 is APLC (Apple Color Matching compatible, ICC v2).
	IccAppleCompatV2 = "applecompat-v2.icc"
	// IccAppleCompatV4 is APLC (Apple Color Matching compatible, ICC v4).
	IccAppleCompatV4 = "applecompat-v4.icc"

	// IccCgats001CompatV2Micro is uCMY (CGATS.001 compatible CMY, compact).
	IccCgats001CompatV2Micro = "cgats001compat-v2-micro.icc"

	// IccColorMatchCompatV2 is ACMC (ColorMatch RGB compatible, ICC v2).
	IccColorMatchCompatV2 = "colormatchcompat-v2.icc"
	// IccColorMatchCompatV4 is ACMC (ColorMatch RGB compatible, ICC v4).
	IccColorMatchCompatV4 = "colormatchcompat-v4.icc"

	// IccDciP3V4 is TP3 (DCI‑P3).
	IccDciP3V4 = "dci-p3-v4.icc"

	// IccDisplayP3V2Magic is sP3 (Display P3, ICC v2 magic).
	IccDisplayP3V2Magic = "displayp3-v2-magic.icc"
	// IccDisplayP3V2Micro is uP3 (Display P3, micro).
	IccDisplayP3V2Micro = "displayp3-v2-micro.icc"
	// IccDisplayP3V4 is sP3 (Display P3, ICC v4).
	IccDisplayP3V4 = "displayp3-v4.icc"

	// IccDisplayP3CompatV2Magic is sP3C (Display P3 compatible, ICC v2 magic).
	IccDisplayP3CompatV2Magic = "displayp3compat-v2-magic.icc"
	// IccDisplayP3CompatV2Micro is uP3C (Display P3 compatible, micro).
	IccDisplayP3CompatV2Micro = "displayp3compat-v2-micro.icc"
	// IccDisplayP3CompatV4 is sP3C (Display P3 compatible, ICC v4).
	IccDisplayP3CompatV4 = "displayp3compat-v4.icc"

	// IccProPhotoV2Magic is uROM (ProPhoto RGB compact).
	IccProPhotoV2Magic = "prophoto-v2-magic.icc"
	// IccProPhotoV2Micro is uROM (ProPhoto RGB micro).
	IccProPhotoV2Micro = "prophoto-v2-micro.icc"
	// IccProPhotoV4 is ROMM (ProPhoto/ROMM RGB, ICC v4).
	IccProPhotoV4 = "prophoto-v4.icc"

	// IccRec2020Gamma24V4 is 2024 (Rec.2020 gamma 2.4, ICC v4).
	IccRec2020Gamma24V4 = "rec2020-g24-v4.icc"
	// IccRec2020V2Magic is 2020 (Rec.2020, ICC v2 magic).
	IccRec2020V2Magic = "rec2020-v2-magic.icc"
	// IccRec2020V2Micro is u202 (Rec.2020 micro).
	IccRec2020V2Micro = "rec2020-v2-micro.icc"
	// IccRec2020V4 is 2020 (Rec.2020, ICC v4).
	IccRec2020V4 = "rec2020-v4.icc"

	// IccRec2020CompatV2Magic is 202C (Rec.2020 compatible, ICC v2 magic).
	IccRec2020CompatV2Magic = "rec2020compat-v2-magic.icc"
	// IccRec2020CompatV2Micro is u20C (Rec.2020 compatible, micro).
	IccRec2020CompatV2Micro = "rec2020compat-v2-micro.icc"
	// IccRec2020CompatV4 is 202C (Rec.2020 compatible, ICC v4).
	IccRec2020CompatV4 = "rec2020compat-v4.icc"

	// IccRec601NtscV2Magic is R601 (Rec.601 NTSC, ICC v2 magic).
	IccRec601NtscV2Magic = "rec601ntsc-v2-magic.icc"
	// IccRec601NtscV2Micro is u601 (Rec.601 NTSC, micro).
	IccRec601NtscV2Micro = "rec601ntsc-v2-micro.icc"
	// IccRec601NtscV4 is R601 (Rec.601 NTSC, ICC v4).
	IccRec601NtscV4 = "rec601ntsc-v4.icc"

	// IccRec601PalV2Magic is 601P (Rec.601 PAL, ICC v2 magic).
	IccRec601PalV2Magic = "rec601pal-v2-magic.icc"
	// IccRec601PalV2Micro is u60P (Rec.601 PAL, micro).
	IccRec601PalV2Micro = "rec601pal-v2-micro.icc"
	// IccRec601PalV4 is 601P (Rec.601 PAL, ICC v4).
	IccRec601PalV4 = "rec601pal-v4.icc"

	// IccRec709V2Magic is R709 (Rec.709, ICC v2 magic).
	IccRec709V2Magic = "rec709-v2-magic.icc"
	// IccRec709V2Micro is u709 (Rec.709, micro).
	IccRec709V2Micro = "rec709-v2-micro.icc"
	// IccRec709V4 is R709 (Rec.709, ICC v4).
	IccRec709V4 = "rec709-v4.icc"

	// IccScRgbV2 is cRGB (scRGB, ICC v2).
	IccScRgbV2 = "scrgb-v2.icc"

	// IccSGreyV2Magic is sGry (Display P3 compatible gray, ICC v2 magic).
	IccSGreyV2Magic = "sgrey-v2-magic.icc"
	// IccSGreyV2Micro is uGry (Display P3 compatible gray, micro).
	IccSGreyV2Micro = "sgrey-v2-micro.icc"
	// IccSGreyV2Nano is nGry (Display P3 compatible gray, nano).
	IccSGreyV2Nano = "sgrey-v2-nano.icc"
	// IccSGreyV4 is sGry (Display P3 compatible gray, ICC v4).
	IccSGreyV4 = "sgrey-v4.icc"

	// IccSRgbV2Magic is sRGB (standard sRGB, ICC v2 magic).
	IccSRgbV2Magic = "srgb-v2-magic.icc"
	// IccSRgbV2Micro is uRGB (sRGB micro).
	IccSRgbV2Micro = "srgb-v2-micro.icc"
	// IccSRgbV2Nano is nRGB (sRGB nano).
	IccSRgbV2Nano = "srgb-v2-nano.icc"
	// IccSRgbV4 is sRGB (standard sRGB, ICC v4).
	IccSRgbV4 = "srgb-v4.icc"

	// IccWideGamutCompatV2 is AWGC (Adobe Wide Gamut compatible, ICC v2).
	IccWideGamutCompatV2 = "widegamutcompat-v2.icc"
	// IccWideGamutCompatV4 is AWGC (Adobe Wide Gamut compatible, ICC v4).
	IccWideGamutCompatV4 = "widegamutcompat-v4.icc"
)

// IccProfiles lists all bundled ICC profile filenames in one place so tests and
// callers can iterate or validate the full set shipped in assets/profiles/icc.
var IccProfiles = []string{
	IccAdobeRGBCompat,
	IccAdobeRGBCompatV2,
	IccAdobeRGBCompatV4,
	IccAppleCompatV2,
	IccAppleCompatV4,
	IccCgats001CompatV2Micro,
	IccColorMatchCompatV2,
	IccColorMatchCompatV4,
	IccDciP3V4,
	IccDisplayP3V2Magic,
	IccDisplayP3V2Micro,
	IccDisplayP3V4,
	IccDisplayP3CompatV2Magic,
	IccDisplayP3CompatV2Micro,
	IccDisplayP3CompatV4,
	IccProPhotoV2Magic,
	IccProPhotoV2Micro,
	IccProPhotoV4,
	IccRec2020Gamma24V4,
	IccRec2020V2Magic,
	IccRec2020V2Micro,
	IccRec2020V4,
	IccRec2020CompatV2Magic,
	IccRec2020CompatV2Micro,
	IccRec2020CompatV4,
	IccRec601NtscV2Magic,
	IccRec601NtscV2Micro,
	IccRec601NtscV4,
	IccRec601PalV2Magic,
	IccRec601PalV2Micro,
	IccRec601PalV4,
	IccRec709V2Magic,
	IccRec709V2Micro,
	IccRec709V4,
	IccScRgbV2,
	IccSGreyV2Magic,
	IccSGreyV2Micro,
	IccSGreyV2Nano,
	IccSGreyV4,
	IccSRgbV2Magic,
	IccSRgbV2Micro,
	IccSRgbV2Nano,
	IccSRgbV4,
	IccWideGamutCompatV2,
	IccWideGamutCompatV4,
}

// GetIccProfile returns the absolute path to the first requested ICC profile
// that is present in assets/profiles/icc. It validates existence so callers
// can embed profiles without risking a panic or missing file error.
func GetIccProfile(profiles ...string) (string, error) {
	if len(profiles) == 0 {
		return "", errors.New("no icc profiles specified")
	}

	// Find first ICC profile file that exists.
	for _, p := range profiles {
		filePath := filepath.Join(IccProfilesPath, p)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			return filePath, nil
		}
	}

	return "", fmt.Errorf("no matching icc profiles found (%s)", strings.Join(profiles, ", "))
}
