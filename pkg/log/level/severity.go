package level

// Severity represents the severity level of an event.
type Severity uint8

// String returns the severity level as a string, e.g. Alert becomes "alert".
func (level Severity) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}
