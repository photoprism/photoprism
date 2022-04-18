package fs

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
)

type FileCodec string

const (
	CodecAVC  FileCodec = "avc1"
	CodecHEVC FileCodec = "hvc1"
	CodecAV1  FileCodec = "av01"
	CodecJpeg FileCodec = "jpeg"
)
