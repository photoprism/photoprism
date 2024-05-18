package thumb

import (
	"strconv"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Standard JPEG image quality levels,
// see https://docs.photoprism.app/user-guide/settings/advanced/#jpeg-quality
const (
	QualityMax    Quality = 90
	QualityHigh   Quality = 85
	QualityMedium Quality = 83
	QualityLow    Quality = 78
	QualityMin    Quality = 70
)

// JpegQualityDefault sets the compression level of newly created JPEGs.
var JpegQualityDefault = QualityMedium

// Quality represents a JPEG image quality.
type Quality int

// EncodeOption returns the quality as imaging.EncodeOption.
func (q Quality) EncodeOption() imaging.EncodeOption {
	return imaging.JPEGQuality(int(q))
}

// String returns the quality as string.
func (q Quality) String() string {
	return strconv.Itoa(int(q))
}

// Int returns the quality as int.
func (q Quality) Int() int {
	return int(q)
}

// QualityLevels maps human-readable settings to a numeric Quality.
var QualityLevels = map[string]Quality{
	"max":      QualityMax,
	"ultra":    QualityMax,
	"best":     QualityMax,
	"6":        QualityMax,
	"5":        QualityMax,
	"high":     QualityHigh,
	"good":     QualityHigh,
	"4":        QualityHigh,
	"medium":   QualityMedium,
	"med":      QualityMedium,
	"default":  QualityMedium,
	"standard": QualityMedium,
	"auto":     QualityMedium,
	"":         QualityMedium,
	"3":        QualityMedium,
	"low":      QualityLow,
	"small":    QualityLow,
	"2":        QualityLow,
	"min":      QualityMin,
	"1":        QualityMin,
	"0":        QualityMin,
}

// JpegQuality returns the JPEG image quality depending on the image size.
func JpegQuality(width, height int) Quality {
	// Use default quality for images larger than 150 pixels.
	if width > 150 || height > 150 {
		return JpegQualityDefault
	}

	// Use lower quality for very small thumbnails.
	return JpegQualitySmall()
}

// JpegQualitySmall returns the quality for images that should be more heavily compressed.
func JpegQualitySmall() Quality {
	if q := JpegQualityDefault - 5; q < QualityMin || q > QualityMax {
		return JpegQualityDefault
	} else {
		return q
	}
}

// ParseQuality returns the matching quality based on a config value string.
func ParseQuality(s string) Quality {
	// Default if empty.
	if s == "" {
		return QualityMedium
	}

	// Try to parse as positive integer.
	if i := txt.Int(s); i >= 25 && i <= 100 {
		return Quality(i)
	}

	// Normalize value.
	s = strings.ToLower(strings.TrimSpace(s))

	// Human-readable quality levels.
	if l, ok := QualityLevels[s]; ok && l > 0 {
		return l
	}

	return QualityMedium
}
