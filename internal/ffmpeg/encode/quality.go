package encode

import (
	"fmt"
)

// Encoding quality min, max, and default settings,
// where 100 is almost lossless.
const (
	BestQuality    = 100
	DefaultQuality = 50
	WorstQuality   = 1
)

// QvQuality returns the video encoding quality as "-q:v" parameter string.
func QvQuality(q int) string {
	switch {
	case q < 0:
		return "50"
	case q < 1:
		return "1"
	case q > 100:
		return "100"
	default:
		return fmt.Sprintf("%d", q)
	}
}

// GlobalQuality returns the video encoding quality as Intel QSV "-global_quality" parameter string.
// The scale is steeper than CRF, so that the default quality matches libx264 at "-crf 25".
func GlobalQuality(q int) string {
	if q <= 0 {
		q = DefaultQuality
	} else if q > BestQuality {
		q = BestQuality
	}

	result := (100-q)*3/5 - 2

	switch {
	case result < 1:
		return "1"
	case result > 51:
		return "51"
	default:
		return fmt.Sprintf("%d", result)
	}
}

// CrfQuality returns the video encoding quality as "-crf" parameter string.
func CrfQuality(q int) string {
	if q <= 0 {
		q = DefaultQuality
	} else if q > BestQuality {
		q = BestQuality
	}

	result := (100 - q) / 2

	switch {
	case result < 1:
		return "0"
	case result > 50:
		return "51"
	default:
		return fmt.Sprintf("%d", result)
	}
}

// QpQuality returns the video encoding quality as "-qp" parameter string.
func QpQuality(q int) string {
	if q <= 0 {
		q = DefaultQuality
	} else if q > BestQuality {
		q = BestQuality
	}

	result := (100 - q) / 2

	switch {
	case result < 1:
		return "0"
	case result > 50:
		return "51"
	default:
		return fmt.Sprintf("%d", result)
	}
}

// CqQuality returns the video encoding quality as NVENC "-cq" parameter string.
// The scale is steeper than CRF, so that the default quality matches libx264 at "-crf 25".
func CqQuality(q int) string {
	if q <= 0 {
		q = DefaultQuality
	} else if q > BestQuality {
		q = BestQuality
	}

	result := 1 + (100-q)*3/5

	switch {
	case result < 1:
		return "1"
	case result > 50:
		return "50"
	default:
		return fmt.Sprintf("%d", result)
	}
}

// VaapiQuality returns the video encoding quality as VA-API "-qp" parameter string.
// It is offset from the CRF scale, so that the default quality matches libx264 at "-crf 25".
func VaapiQuality(q int) string {
	if q <= 0 {
		q = DefaultQuality
	} else if q > BestQuality {
		q = BestQuality
	}

	result := (100-q)/2 + 4

	if result > 51 {
		return "51"
	}

	return fmt.Sprintf("%d", result)
}
