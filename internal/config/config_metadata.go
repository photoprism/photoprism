package config

import "time"

// ConvertTimeout returns the budget for converting one still image, document, or RAW file,
// or 0 when no limit applies. The budget covers the whole file: the converters tried for it
// share it, so the time a single file can occupy does not grow with the number of candidates.
func (c *Config) ConvertTimeout() time.Duration {
	return convertBudget(c.options.ConvertTimeout, DefaultConvertTimeout)
}

// TranscodeTimeout returns the budget for transcoding one video, or 0 when no limit applies.
func (c *Config) TranscodeTimeout() time.Duration {
	return convertBudget(c.options.TranscodeTimeout, DefaultTranscodeTimeout)
}

// convertBudget turns a configured number of minutes into a duration, applying the given
// default when unset and reporting 0 for a negative value, which means no limit.
func convertBudget(minutes, fallback int) time.Duration {
	if minutes == 0 {
		minutes = fallback
	}

	switch {
	case minutes < 0:
		return 0
	case minutes > MaxConvertTimeout:
		minutes = MaxConvertTimeout
	}

	return time.Duration(minutes) * time.Minute
}

// ExifBruteForce checks if a brute-force search should be performed when no Exif headers were found.
func (c *Config) ExifBruteForce() bool {
	return c.options.ExifBruteForce || !c.ExifToolJson()
}

// ExifToolBin returns the exiftool executable file name.
func (c *Config) ExifToolBin() string {
	return FindBin(c.options.ExifToolBin, "exiftool")
}

// ExifToolJson checks if creating JSON metadata sidecar files with Exiftool is enabled.
func (c *Config) ExifToolJson() bool {
	return !c.DisableExifTool()
}
