package photoprism

func NewTestConfig() *Config {
	return &Config{
		DarktableCli:   "/Applications/darktable.app/Contents/MacOS/darktable-cli",
		OriginalsPath:  GetExpandedFilename("photos/originals"),
		ThumbnailsPath: GetExpandedFilename("photos/thumbnails"),
		ImportPath:     GetExpandedFilename("photos/import"),
		ExportPath:     GetExpandedFilename("photos/export"),
	}
}
