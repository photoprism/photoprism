package thumb

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// Package configuration variables.
var (
	Library            = LibImaging        // Image processing library to be used.
	Color              = ColorAuto         // Color sets the standard color profile for thumbnails.
	Filter             = ResampleLanczos   // Filter specifies the default downscaling filter.
	SizeCached         = SizeFit1920.Width // Pre-generated thumbnail size limit.
	SizeOnDemand       = SizeFit5120.Width // On-demand thumbnail size limit.
	JpegQualityDefault = QualityMedium     // JpegQualityDefault sets the compression level of newly created JPEGs.
	CachePublic        = false             // Specifies if static content may be cached by a CDN or caching proxy.
	ExamplesPath       = fs.Abs("../../assets/examples")
	IccProfilesPath    = fs.Abs("../../assets/profiles/icc")
)
