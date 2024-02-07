package ffmpeg

import "fmt"

// Options represents transcoding options.
type Options struct {
	Bin      string
	Encoder  AvcEncoder
	Size     int
	Bitrate  string
	MapVideo string
	MapAudio string
}

// VideoFilter returns the FFmpeg video filter string based on the size limit in pixels and the pixel format.
func (o Options) VideoFilter(format PixelFormat) string {
	// scale specifies the FFmpeg downscale filter, see http://trac.ffmpeg.org/wiki/Scaling.
	if format == "" {
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))'", o.Size, o.Size)
	} else if format == FormatQSV {
		return fmt.Sprintf("vpp_qsv=framerate=30,scale_qsv=w='if(gte(iw,ih), min(%d, iw), -1)':h='if(gte(iw,ih), -1, min(%d, ih))'", o.Size, o.Size)
	} else {
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))',format=%s", o.Size, o.Size, format)
	}
}
