package fs

// TypeAnimated maps animated file types to their mime type.
var TypeAnimated = TypeMap{
	ImageGIF:   MimeTypeGIF,
	ImagePNG:   MimeTypeAPNG,
	ImageWebP:  MimeTypeWebP,
	ImageAVIF:  MimeTypeAVIFS,
	ImageAVIFS: MimeTypeAVIFS,
	ImageHEIC:  MimeTypeHEICS,
	ImageHEICS: MimeTypeHEICS,
}
