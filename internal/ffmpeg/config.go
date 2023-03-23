package ffmpeg

// Options represents transcoding options.
type Options struct {
	Bin      string
	Encoder  AvcEncoder
	Bitrate  string
	MapVideo string
	MapAudio string
}
