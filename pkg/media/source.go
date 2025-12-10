package media

// Src identifies a media source.
type Src = string

// Data source types.
const (
	// SrcLocal indicates the media originates from local storage.
	SrcLocal Src = "local"
	// SrcRemote indicates the media originates from a remote source.
	SrcRemote Src = "remote"
)
