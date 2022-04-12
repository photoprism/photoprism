package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// Format represents a video format standard.
type Format struct {
	File   fs.Format
	Codec  Codec
	Width  int
	Height int
	Public bool
}

// FormatNames maps names to video format standards.
type FormatNames map[string]Format

var MP4 = Format{
	File:   fs.FormatMp4,
	Codec:  CodecAVC,
	Width:  0,
	Height: 0,
	Public: true,
}

var AVC = Format{
	File:   fs.FormatAVC,
	Codec:  CodecAVC,
	Width:  0,
	Height: 0,
	Public: true,
}

var AV1 = Format{
	File:   fs.FormatAV1,
	Codec:  CodecAV1,
	Width:  0,
	Height: 0,
	Public: false,
}

var HEVC = Format{
	File:   fs.FormatHEVC,
	Codec:  CodecHEVC,
	Width:  0,
	Height: 0,
	Public: false,
}

var Formats = FormatNames{
	"":     AVC,
	"mp4":  MP4,
	"avc":  AVC,
	"av1":  AV1,
	"hevc": HEVC,
}
