package fs

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
)

type FileCodec string

const (
	CodecAvc   FileCodec = "avc1"
	CodecHvc   FileCodec = "hvc1"
	CodecJpeg  FileCodec = "jpeg"
	CodecOther FileCodec = ""
)
