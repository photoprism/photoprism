package config

// ExifBruteForce checks if a brute-force search should be performed when no Exif headers were found.
func (c *Config) ExifBruteForce() bool {
	return c.options.ExifBruteForce || !c.ExifToolJson()
}

// ExifToolBin returns the exiftool executable file name.
func (c *Config) ExifToolBin() string {
	return findBin(c.options.ExifToolBin, "exiftool")
}

// ExifToolJson checks if creating JSON metadata sidecar files with Exiftool is enabled.
func (c *Config) ExifToolJson() bool {
	return !c.DisableExifTool()
}
