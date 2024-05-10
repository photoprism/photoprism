package thumb

import (
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/disintegration/imaging"
)

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

// Common Quality levels.
// see https://docs.photoprism.app/user-guide/settings/advanced/#jpeg-quality
const (
	QualityBest    Quality = 95
	QualityHigh    Quality = 88
	QualityDefault Quality = 82
	QualityLow     Quality = 80
	QualityBad     Quality = 75
	QualityWorst   Quality = 70
)

// QualityLevels maps human-readable settings to a numeric Quality.
var QualityLevels = map[string]Quality{
	"5":         QualityBest,
	"ultra":     QualityBest,
	"best":      QualityBest,
	"4":         QualityHigh,
	"excellent": QualityHigh,
	"good":      QualityHigh,
	"high":      QualityHigh,
	"3":         QualityDefault,
	"":          QualityDefault,
	"ok":        QualityDefault,
	"default":   QualityDefault,
	"standard":  QualityDefault,
	"medium":    QualityDefault,
	"2":         QualityLow,
	"low":       QualityLow,
	"small":     QualityLow,
	"1":         QualityBad,
	"bad":       QualityBad,
	"0":         QualityWorst,
	"worst":     QualityWorst,
	"lowest":    QualityWorst,
}

// Current Quality settings.
var (
	JpegQuality      = QualityDefault
	JpegQualitySmall = QualityLow
)

// ParseQuality returns the matching quality based on a config value string.
func ParseQuality(s string) Quality {
	// Default if empty.
	if s == "" {
		return QualityDefault
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

	return QualityDefault
}
