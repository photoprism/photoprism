package clean

const (
	// LengthType is the default max length for numeric types.
	LengthType = 64
	// LengthShortType limits short numeric values.
	LengthShortType = 8
	// LengthIPv6 defines the maximum IPv6 string length.
	LengthIPv6 = 39
	// LengthLog caps log messages to this length.
	LengthLog = 512
	// LengthLimit is the global hard limit for sanitized strings.
	LengthLimit = 4096
)
